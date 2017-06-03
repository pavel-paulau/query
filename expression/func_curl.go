//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package expression

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"

	curl "github.com/andelf/go-curl"
	"github.com/couchbase/query/auth"
	"github.com/couchbase/query/util"
	"github.com/couchbase/query/value"

	"github.com/couchbase/query/logging"
)

///////////////////////////////////////////////////
//
// Curl
//
///////////////////////////////////////////////////

// To look at values for headers see https://sourceforge.net/p/curl/bugs/385/
// For a full list see :
// https://github.com/curl/curl/blob/6b7616690e5370c21e3a760321af6bf4edbabfb6/include/curl/curl.h

// Protocol constants
const (
	_CURLPROTO_HTTP  = 1 << 0 /* HTTP Protocol */
	_CURLPROTO_HTTPS = 1 << 1 /* HTTPS Protocol */

)

// Authentication constants
const (
	_CURLAUTH_BASIC = 1 << 0 /* Basic (default)*/
	_CURLAUTH_ANY   = ^(0)   /* all types set */
)

// N1QL User-Agent value
const (
	_N1QL_USER_AGENT = "couchbase/n1ql/" + util.VERSION
)

// Max request size from server (cant import because of cyclic dependency)
const (
	MIN_RESPONSE_SIZE = 20 * (1 << 20)
	MAX_RESPONSE_SIZE = 64 * (1 << 20)
)

// Path to certs and whitelist
const (
	_PATH = "/../var/lib/couchbase/n1qlcerts/"
)

var hostname string

/*
This represents the curl function CURL(method, url, options).
It returns result of the curl operation on the url based on
the method and options.
*/
type Curl struct {
	FunctionBase
	myCurl *curl.CURL
}

func NewCurl(operands ...Expression) Function {
	rv := &Curl{
		*NewFunctionBase("curl", operands...),
		nil,
	}

	rv.volatile = true
	rv.expr = rv
	return rv
}

/*
Visitor pattern.
*/
func (this *Curl) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitFunction(this)
}

func (this *Curl) Type() value.Type { return value.OBJECT }

func (this *Curl) Evaluate(item value.Value, context Context) (value.Value, error) {
	return this.Eval(this, item, context)
}

func (this *Curl) Privileges() *auth.Privileges {
	unionPrivileges := auth.NewPrivileges()
	unionPrivileges.Add("", auth.PRIV_QUERY_EXTERNAL_ACCESS)

	children := this.Children()
	for _, child := range children {
		unionPrivileges.AddAll(child.Privileges())
	}

	return unionPrivileges
}

func (this *Curl) Apply(context Context, args ...value.Value) (value.Value, error) {

	// Get ip addresses to display in error

	name, _ := os.Hostname()

	addrs, err := net.LookupHost(name)

	if err != nil {
		logging.Infof("Error looking up hostname: %v\n", err)
	}

	hostname = strings.Join(addrs, ",")

	// In order to have restricted access, the administrator will have to create
	// curl_whitelist.json with the all_access field set to false.
	// In order to access all endpoints, the administrator will have to create
	// curl_whitelist.json with the all_access field set to true.

	// Before performing any checks, see if curl_whitelist.json exists.
	// 1. If it does not exist, then return with error. (Disable access to the CURL function)
	// 2. For all other cases, CURL can execute depending on contents of the file, but we defer
	//    whitelist check to handle_curl()

	errInit := initialCheck()

	if errInit != nil {
		// This means that we dont have access to the curl funtion. Case 1 from above.
		return value.NULL_VALUE, errInit
	}

	if this.myCurl == nil {
		this.myCurl = curl.EasyInit()
		if this.myCurl == nil {
			return value.NULL_VALUE, fmt.Errorf("Error initializing libcurl")
		}
	}
	// End libcurl easy session
	defer func() {
		if this.myCurl != nil {
			this.myCurl.Cleanup()
			this.myCurl = nil
		}
	}()

	// URL
	first := args[0]
	if first.Type() == value.MISSING {
		return value.MISSING_VALUE, nil
	} else if first.Type() != value.STRING {
		return value.NULL_VALUE, nil
	}

	// CURL URL
	curl_url := first.Actual().(string)

	// Empty options to pass into curl.
	options := map[string]interface{}{}

	// If we have options then process them.
	if len(args) == 2 {
		second := args[1]

		if second.Type() == value.MISSING {
			return value.MISSING_VALUE, nil
		} else if second.Type() == value.OBJECT {
			//Process the options
			options = second.Actual().(map[string]interface{})
		} else {
			return value.NULL_VALUE, nil
		}
	}

	// Now you have the URL and the options with which to call curl.
	result, err := this.handleCurl(curl_url, options)

	if err != nil {
		return value.NULL_VALUE, err
	}

	// For Silent mode where we dont want any output.
	switch results := result.(type) {
	case map[string]interface{}:
		if len(results) == 0 {
			return value.MISSING_VALUE, nil
		}
	case []interface{}:
		if len(results) == 0 {
			return value.MISSING_VALUE, nil
		}

	default:
		return value.NULL_VALUE, nil
	}

	return value.NewValue(result), nil
}

func (this *Curl) Indexable() bool {
	return false
}

func (this *Curl) MinArgs() int { return 1 }

func (this *Curl) MaxArgs() int { return 2 }

/*
Factory method pattern.
*/
func (this *Curl) Constructor() FunctionConstructor {
	return NewCurl
}

func (this *Curl) handleCurl(url string, options map[string]interface{}) (interface{}, error) {
	// Handle different cases

	// initial check for curl_whitelist.json has been completed. The file exists.
	// Now we need to access the contents of the file and check for validity.
	err := whitelistCheck(url)
	if err != nil {
		return nil, err
	}

	// For result-cap and request size
	responseSize := setResponseSize(MIN_RESPONSE_SIZE)
	sizeError := false

	// For data method
	getMethod := false
	dataOp := false
	stringData := ""
	encodedData := false

	// For silent mode
	silent := false

	// To show errors encountered when executing the CURL function.
	show_error := true

	showErrVal, ok := options["show_error"]
	if ok {
		if value.NewValue(showErrVal).Type() != value.BOOLEAN {
			return nil, fmt.Errorf(" Incorrect type for show_error option in CURL ")
		}
		show_error = value.NewValue(showErrVal).Actual().(bool)
	}

	// Set MAX_REDIRS to 0 as an added precaution to disable redirection.
	/*
		Libcurl code to set MAX_REDIRS
		curl_easy_setopt(hnd, CURLOPT_MAXREDIRS, 50L);
	*/
	this.myCurl.Setopt(curl.OPT_MAXREDIRS, 0)

	// Set what protocols are allowed.
	/*
		CURL.H  CURLPROTO_ defines are for the CURLOPT_*PROTOCOLS options
		#define CURLPROTO_HTTP   (1<<0)
		#define CURLPROTO_HTTPS  (1<<1)

		Libcurl code to set what protocols are allowed.
		curl_easy_setopt(curl, CURLOPT_PROTOCOLS,CURLPROTO_HTTP | CURLPROTO_HTTPS);
	*/
	this.myCurl.Setopt(curl.OPT_PROTOCOLS, _CURLPROTO_HTTP|_CURLPROTO_HTTPS)

	// Prepare a header []string - slist1 as per libcurl.
	header := []string{}

	// Set curl User-Agent by default.
	this.curlUserAgent(_N1QL_USER_AGENT)

	// When we dont have options, but only have the URL.
	/*
		Libcurl code to set the url
		curl_easy_setopt(hnd, CURLOPT_URL, "https://api.github.com/users/ikandaswamy/repos");
	*/
	this.myCurl.Setopt(curl.OPT_URL, url)

	for k, val := range options {
		// Only support valid options.
		switch k {
		/*
			show_error: Do not output the errors with the CURL function
			in case this is set. This is handled in the beginning.
		*/
		case "show-error", "--show-error", "S", "-S":
			break
		/*
			get: Send the -d data with a HTTP GET (H)
			Since we set the curl method as the first argument, it is
			important to note that providing this option does nothing.
		*/
		case "get", "--get", "G", "-G":
			if value.NewValue(val).Type() != value.BOOLEAN {
				if show_error == true {
					return nil, fmt.Errorf(" Incorrect type for get option in CURL ")
				} else {
					return nil, nil
				}
			}
			get := value.NewValue(val).Actual().(bool)
			if get {
				getMethod = true
				this.simpleGet(url)
			}

		/*
		   request: Specify request method to use. Since we set
		   the curl method as the first argument, it is important
		   to note that providing this option does nothing.
		*/
		case "request", "--request", "X", "-X":
			request := value.NewValue(val)
			if request.Type() != value.STRING {
				return nil, fmt.Errorf(" Incorrect type for request option in CURL. It should be a string. ")
			}

			// Remove the quotations at the end.
			requestVal := request.String()
			requestVal = requestVal[1 : len(requestVal)-1]

			// Methods are case sensitive.
			if requestVal != "GET" && requestVal != "POST" {
				if show_error == true {
					return nil, fmt.Errorf(" CURL only supports GET and POST requests. ")
				} else {
					return nil, nil
				}
			}

			if requestVal == "GET" {
				getMethod = true
			}

			/*
				Libcurl code to handle requests is
				curl_easy_setopt(hnd, CURLOPT_CUSTOMREQUEST, "POST");
			*/
			this.myCurl.Setopt(curl.OPT_CUSTOMREQUEST, requestVal)

		/*
			data: HTTP POST data (H). However in some cases in CURL
			this can be issued with a GET as well. In these cases, the
			data is appended to the URL followed by a ?.
		*/
		case "data", "--data", "d", "-d", "data-urlencode", "--data-urlencode":

			if k == "data-urlencode" || k == "--data-urlencode" {
				encodedData = true
			}
			dataVal := value.NewValue(val).Actual()

			switch dataVal.(type) {
			case []interface{}:
			case string:
				dataVal = []interface{}{dataVal}
			default:
				if show_error == true {
					return nil, fmt.Errorf(" Incorrect type for data option in CURL.It needs to be a string. ")
				} else {
					return nil, nil
				}
			}

			for _, data := range dataVal.([]interface{}) {
				newDval := value.NewValue(data)
				if newDval.Type() != value.STRING {
					if show_error == true {
						return nil, fmt.Errorf(" Incorrect type for data option. ")
					} else {
						return nil, nil
					}
				}

				dataT := newDval.Actual().(string)

				// If the option is data-urlencode then encode the data first.
				if encodedData {
					// When we encode strings, = should not be encoded.
					// The curl.Escape() method for go behaves different to the libcurl method.
					// q=select 1 should be q=select%201 and not q%3Dselect%201
					// Hence split the string, encode and then rejoin.
					stringComponent := strings.Split(dataT, "=")
					for i, _ := range stringComponent {
						stringComponent[i] = this.myCurl.Escape(stringComponent[i])
					}

					dataT = strings.Join(stringComponent, "=")
				}

				if stringData == "" {
					stringData = dataT
				} else {
					stringData = stringData + "&" + dataT
				}

			}
			dataOp = true

		/*
			header: Pass custom header to server (H). It has to be a string,
			otherwise we error out.
		*/
		case "headers", "header", "--header", "--headers", "H", "-H":
			/*
				Libcurl code to handle multiple headers using the --header and -H options.

				  struct curl_slist *slist1;
				  slist1 = NULL;
				  slist1 = curl_slist_append(slist1, "X-N1QL-User-Agent: couchbase/n1ql/1.7.0");
				  slist1 = curl_slist_append(slist1, "User-Agent: ikandaswamy");
			*/
			// Get the value
			headerVal := value.NewValue(val).Actual()

			switch headerVal.(type) {

			case []interface{}:
				//Do nothing
			case string:
				headerVal = []interface{}{headerVal}

			default:
				if show_error == true {
					return nil, fmt.Errorf(" Incorrect type for header option " + value.NewValue(val).String() + " in CURL. Header option should be a string value or an array of strings.  ")
				} else {
					return nil, nil
				}
			}

			// We have an array of interfaces that represent different fields in the Header.
			// Add all the headers to a []string to pass to OPT_HTTPHEADER
			for _, hval := range headerVal.([]interface{}) {

				newHval := value.NewValue(hval)
				if newHval.Type() != value.STRING {
					if show_error == true {
						return nil, fmt.Errorf(" Incorrect type for header option " + newHval.String() + " in CURL. Header option should be a string value or an array of strings.  ")
					} else {
						return nil, nil
					}

				}
				head := newHval.String()
				header = append(header, head[1:len(head)-1])
			}

		/*
			silent: Do not output anything. It has to be a boolean, otherwise
			we error out.
		*/
		case "silent", "--silent", "s", "-s":
			if value.NewValue(val).Type() != value.BOOLEAN {
				if show_error == true {
					return nil, fmt.Errorf(" Incorrect type for silent option in CURL ")
				} else {
					return nil, nil
				}
			}
			silent = value.NewValue(val).Actual().(bool)
		/*
			connect-timeout: Maximum time allowed for connection in seconds
		*/
		case "connect-timeout", "--connect-timeout":
			/*
				Libcurl code to set connect-timeout is
				curl_easy_setopt(hnd, CURLOPT_CONNECTTIMEOUT_MS, 1000L);

				To save fractions of the decimal value, libcurl uses the _MS suffix to convert
				to milliseconds.
			*/
			if value.NewValue(val).Type() != value.NUMBER {
				return nil, fmt.Errorf(" Incorrect type for connect-timeout option in CURL ")
			}

			connTime := value.NewValue(val).Actual().(float64)

			this.curlConnectTimeout(int64(connTime))
		/*
			max-time: Maximum time allowed for the transfer in seconds
		*/
		case "max-time", "--max-time", "m", "-m":
			/*
				Libcurl code to set max-time is
				curl_easy_setopt(hnd, CURLOPT_TIMEOUT_MS, 1000L);

				To save fractions of the decimal value, libcurl uses the _MS suffix to convert
				to milliseconds.
			*/
			if value.NewValue(val).Type() != value.NUMBER {
				return nil, fmt.Errorf(" Incorrect type for max-time option in CURL ")
			}

			maxTime := value.NewValue(val).Actual().(float64)

			this.curlMaxTime(int64(maxTime))
		/*
			user: Server user and password separated by a :. By default if a
			password is not specified, then use an empty password.
		*/
		case "user", "--user", "-u", "u":
			/*
				Libcurl code to set user
				curl_easy_setopt(hnd, CURLOPT_USERPWD, "Administrator:password");
			*/
			if value.NewValue(val).Type() != value.STRING {
				return nil, fmt.Errorf(" Incorrect type for user option in CURL. It should be a string. ")
			}
			this.curlAuth(value.NewValue(val).String())
		/*
			basic: Use HTTP Basic Authentication. It has to be a boolean, otherwise
			we error out.
		*/
		case "basic", "--basic":
			/*
				Libcurl code to set --basic
				#define CURLAUTH_BASIC (1<<0) /* Basic (default)
				curl_easy_setopt(hnd, CURLOPT_HTTPAUTH, (long)CURLAUTH_BASIC);
			*/

			if value.NewValue(val).Type() != value.BOOLEAN {
				if show_error == true {
					return nil, fmt.Errorf(" Incorrect type for basic option in CURL ")
				} else {
					return nil, nil
				}
			}
			if value.NewValue(val).Actual().(bool) == true {
				this.myCurl.Setopt(curl.OPT_HTTPAUTH, _CURLAUTH_BASIC)
			}
		/*
			anyauth: curl to figure out authentication method by itself, and use the most secure one.
			It has to be a boolean, otherwise we error out.
		*/
		case "anyauth", "--anyauth":
			/*
				Libcurl code to set --anyauth
				#define CURLAUTH_ANY ~0
				curl_easy_setopt(hnd, CURLOPT_HTTPAUTH, (long)CURLAUTH_ANY);
			*/
			if value.NewValue(val).Type() != value.BOOLEAN {
				if show_error == true {
					return nil, fmt.Errorf(" Incorrect type for anyauth option in CURL ")
				} else {
					return nil, nil
				}
			}
			if value.NewValue(val).Actual().(bool) == true {
				this.myCurl.Setopt(curl.OPT_HTTPAUTH, _CURLAUTH_ANY)
			}
		/*
			insecure: Allow connections to SSL sites without certs (H). It has to be a boolean,
			otherwise we error out.
		*/
		case "insecure", "--insecure", "k", "-k":
			/*
				Set the value to 1 for strict certificate check please
				curl_easy_setopt(curl, CURLOPT_SSL_VERIFYPEER, 1L);

				If you want to connect to a site who isn't using a certificate that is
				signed by one of the certs in the CA bundle you have, you can skip the
				verification of the server's certificate. This makes the connection
				A LOT LESS SECURE.
			*/
			if value.NewValue(val).Type() != value.BOOLEAN {
				if show_error == true {
					return nil, fmt.Errorf(" Incorrect type for insecure option in CURL ")
				} else {
					return nil, nil
				}
			}
			insecure := value.NewValue(val).Actual().(bool)
			if insecure == true {
				this.myCurl.Setopt(curl.OPT_SSL_VERIFYPEER, insecure)
			}
		/*
			keepalive-time: Wait SECONDS between keepalive probes for low level TCP connectivity.
			(Does not affect HTTP level keep-alive)

		*/
		case "keepalive-time", "--keepalive-time":
			/*
				Libcurl code to set keepalive-time
				curl_easy_setopt(hnd, CURLOPT_TCP_KEEPALIVE, 1L);
				curl_easy_setopt(hnd, CURLOPT_TCP_KEEPIDLE, 1L);
				curl_easy_setopt(hnd, CURLOPT_TCP_KEEPINTVL, 1L);
			*/
			if value.NewValue(val).Type() != value.NUMBER {
				return nil, fmt.Errorf(" Incorrect type for keepalive-time option in CURL ")
			}

			kATime := value.NewValue(val).Actual().(float64)

			this.curlKeepAlive(int64(kATime))

		/*
			user-agent: Value for the User-Agent to send to the server.
		*/
		case "user-agent", "--user-agent", "A", "-A":
			/*
				Libcurl code to set user-agent
				curl_easy_setopt(hnd, CURLOPT_USERAGENT, "curl/7.43.0");
			*/
			if value.NewValue(val).Type() != value.STRING {
				return nil, fmt.Errorf(" Incorrect type for user-agent option in CURL. user-agent should be a string. ")
			}
			userAgent := value.NewValue(val).Actual().(string)
			this.curlUserAgent(userAgent)

		case "cacert":
			/*
				Cert is stored PEM coded in file.
				curl_easy_setopt(curl, CURLOPT_SSLCERTTYPE, "PEM");
				curl_easy_setopt(hnd, CURLOPT_CAINFO, "ca.pem");
			*/
			// All the certificates are stored withing the ..var/lib/couchbase/n1qlcerts
			// Find the os
			subdir := filepath.FromSlash(_PATH)

			// Get directory of currently running file.
			certDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
			if err != nil {
				return nil, fmt.Errorf(" ../var/lib/couchbase/n1qlcerts/ does not exist on node " + hostname)
			}

			// nsserver uses the inbox folder within var/lib/couchbase to read certificates from.
			certDir = certDir + subdir
			// dir. Paths are not allowed.
			if value.NewValue(val).Type() != value.STRING {
				return nil, fmt.Errorf(" Incorrect type for cacert option in CURL. It should be a string. ")
			}
			certName := value.NewValue(val).Actual().(string)

			// Make sure this file is not a path.
			// use path.Split and check 1st and 2nd args

			dir, file := path.Split(certName)
			if dir != "" || file == "" {
				return nil, fmt.Errorf(" Cacert should only contain the certificate name. Paths are invalid. ")
			}

			// Also make sure the extension is .pem
			if path.Ext(file) != ".pem" {
				return nil, fmt.Errorf(" Cacert should only contain the certificate name that refers to a valid pem file. ")
			}

			this.curlCacert(certDir + file)

		case "result-cap":
			// In order to restrict size of response use curlopt-range.
			// Min allowed = 20MB  20971520
			// Max allowed = request-size-cap default 67 108 864

			if value.NewValue(val).Type() != value.NUMBER {
				return nil, fmt.Errorf(" Incorrect type for result-cap option in CURL ")
			}

			maxSize := value.NewValue(val).Actual().(float64)

			responseSize = setResponseSize(int64(maxSize))

		default:
			return nil, fmt.Errorf(" CURL option %v is not supported.", k)

		}

	}

	/*
		Check if we set the request method to GET either by passing in --get or
		by saying -XGET. This will be used to decide how data is passed for the
		-data option.
	*/
	if dataOp {
		if getMethod {
			this.simpleGet(url + "?" + stringData)
		} else {
			this.curlData(stringData)
		}
	}

	/*
		Libcurl code to write data to chunk of memory
		1. Send all data to this function
		 curl_easy_setopt(curl_handle, CURLOPT_WRITEFUNCTION, writeToBufferFunc);

		2. Pass the chunk to the callback function
		 curl_easy_setopt(curl_handle, CURLOPT_WRITEDATA, (void *)&b);

		3. Define callback function - getinmemory.c example (https://curl.haxx.se/libcurl/c/getinmemory.html)
				static size_t
				WriteMemoryCallback(void *contents, size_t size, size_t nmemb, void *userp)
				{
		  			size_t realsize = size * nmemb;
		  			struct MemoryStruct *mem = (struct MemoryStruct *)userp;

		  			mem->memory = realloc(mem->memory, mem->size + realsize + 1);
		  			if(mem->memory == NULL) {
		    			// out of memory!
		    			printf("not enough memory (realloc returned NULL)\n");
		    			return 0;
		  			}

		 			memcpy(&(mem->memory[mem->size]), contents, realsize);
		 	 		mem->size += realsize;
		  			mem->memory[mem->size] = 0;

		  			return realsize;
				}

		We use the bytes.Buffer package Write method. go-curl fixes the input and output format
		of the callback function to be func(buf []byte, userdata interface{}) bool {}
	*/

	// Set the header, so that the entire []string are passed in.
	this.curlHeader(header)
	this.curlCiphers()

	var b bytes.Buffer

	// Callback function to save data instead of redirecting it into stdout.
	writeToBufferFunc := func(buf []byte, userdata interface{}) bool {
		if silent == false {

			// Check length of buffer b. If it is greater than
			if int64(b.Len()) > responseSize {
				// No more writing we are all done
				// If this interrupts the stream of data then we throw not a JSON endpoint error.
				sizeError = true
				return true
			} else {
				b.Write([]byte(buf))
			}
		}
		return true
	}

	this.myCurl.Setopt(curl.OPT_WRITEFUNCTION, writeToBufferFunc)

	this.myCurl.Setopt(curl.OPT_WRITEDATA, b)

	if err := this.myCurl.Perform(); err != nil {
		if show_error == true {
			return nil, err
		} else {
			return nil, nil
		}
	}

	if sizeError {
		return nil, fmt.Errorf("Response Size has been exceeded. The max response capacity is %v", responseSize)
	}

	// The return type can either be and ARRAY or an OBJECT
	if b.Len() != 0 {
		var dat interface{}

		if err := json.Unmarshal(b.Bytes(), &dat); err != nil {
			if show_error == true {
				return nil, fmt.Errorf("Invalid JSON endpoint %v", url)
			} else {
				return nil, nil
			}
		}

		return dat, nil
	}

	return nil, nil

}

func (this *Curl) simpleGet(url string) {
	myCurl := this.myCurl
	myCurl.Setopt(curl.OPT_URL, url)
	myCurl.Setopt(curl.OPT_HTTPGET, 1)
}

func (this *Curl) curlData(data string) {
	myCurl := this.myCurl
	myCurl.Setopt(curl.OPT_POST, true)
	myCurl.Setopt(curl.OPT_POSTFIELDS, data)
}

func (this *Curl) curlHeader(header []string) {

	/*
		Libcurl code to handle multiple headers using the --header and -H options.
		 slist1 = curl_slist_append(slist1, "X-N1QL-Header: n1ql-1.7.0");
		 curl_easy_setopt(hnd, CURLOPT_HTTPHEADER, slist1);
	*/

	// Set the Custom N1QL Header first.
	// This will allow localhost endpoints to recognize the query service.
	header = append(header, "X-N1QL-User-Agent: "+_N1QL_USER_AGENT)
	myCurl := this.myCurl
	myCurl.Setopt(curl.OPT_HTTPHEADER, header)
}

func (this *Curl) curlUserAgent(userAgent string) {
	/*
		Libcurl code to set user-agent
		curl_easy_setopt(hnd, CURLOPT_USERAGENT, "curl/7.43.0");
	*/
	myCurl := this.myCurl
	myCurl.Setopt(curl.OPT_USERAGENT, userAgent)
}

func (this *Curl) curlAuth(val string) {
	/*
		Libcurl code to set username password
		curl_easy_setopt(hnd, CURLOPT_USERPWD, "Administrator:password");
	*/
	myCurl := this.myCurl
	if val == "" {
		myCurl.Setopt(curl.OPT_USERPWD, "")
	} else {
		val = val[1 : len(val)-1]
		if !strings.Contains(val, ":") {
			// Append an empty password if there isnt one
			val = val + ":" + ""
		}

		myCurl.Setopt(curl.OPT_USERPWD, val)
	}
}

func (this *Curl) curlConnectTimeout(val int64) {
	/*
		Libcurl code to set connect-timeout is
		curl_easy_setopt(hnd, CURLOPT_CONNECTTIMEOUT_MS, 1000L);

		To save fractions of the decimal value, libcurl uses the _MS suffix to convert
		to milliseconds.
	*/
	myCurl := this.myCurl
	myCurl.Setopt(curl.OPT_CONNECTTIMEOUT, val)

}

func (this *Curl) curlMaxTime(val int64) {
	/*
		Libcurl code to set max-time is
		curl_easy_setopt(hnd, CURLOPT_TTIMEOUT_MS, 1000L);

		To save fractions of the decimal value, libcurl uses the _MS suffix to convert
		to milliseconds.
	*/
	myCurl := this.myCurl
	myCurl.Setopt(curl.OPT_TIMEOUT, val)
}

func (this *Curl) curlKeepAlive(val int64) {
	/*
		Libcurl code to set keepalive-time
		curl_easy_setopt(hnd, CURLOPT_TCP_KEEPALIVE, 1L);
		curl_easy_setopt(hnd, CURLOPT_TCP_KEEPIDLE, 1L);
		curl_easy_setopt(hnd, CURLOPT_TCP_KEEPINTVL, 1L);
	*/
	myCurl := this.myCurl
	myCurl.Setopt(curl.OPT_TCP_KEEPALIVE, 1)
	myCurl.Setopt(curl.OPT_TCP_KEEPIDLE, val)
	myCurl.Setopt(curl.OPT_TCP_KEEPINTVL, val)
}

func (this *Curl) curlCacert(fileName string) {
	/*
		Cert is stored PEM coded in file.
		curl_easy_setopt(curl, CURLOPT_SSLCERTTYPE, "PEM");
		curl_easy_setopt(hnd, CURLOPT_CAINFO, "ca.pem");
	*/
	myCurl := this.myCurl
	myCurl.Setopt(curl.OPT_SSLCERTTYPE, "PEM")
	myCurl.Setopt(curl.OPT_CAINFO, fileName)
}

func (this *Curl) curlCiphers() {

	// For the mapping for nss http://unix.stackexchange.com/questions/208437/how-to-convert-ssl-ciphers-to-curl-format
	// For the mapping for openssl (default) - https://wiki.openssl.org/index.php/Manual:Ciphers(1)
	// Each cipher is encoded as a string in curl.
	// The following map gives the mapping from the standard id to the curl specific string representing
	// that cipher.
	ciphersMapping := map[uint16]string{
		tls.TLS_RSA_WITH_AES_128_CBC_SHA:            "AES128-SHA",
		tls.TLS_RSA_WITH_AES_256_CBC_SHA:            "AES256-SHA",
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA:    "ECDHE-ECDSA-AES256-SHA",
		tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA:      "ECDHE-RSA-AES256-SHA",
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256:   "ECDHE-RSA-AES128-GCM-SHA256",
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256: "ECDHE-ECDSA-AES128-GCM-SHA256",
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA:    "ECDHE-ECDSA-AES128-SHA",
		tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA:      "ECDHE-RSA-AES128-SHA",
	}

	// Get the Ciphers supported by couchbase based on the level set
	cbCiphers := util.CipherSuites()

	// Create a comma separated list of the ciphers that need to be used.
	finalCipherList := ""
	for _, cipherId := range cbCiphers {
		if finalCipherList == "" {
			finalCipherList = ciphersMapping[cipherId]
		} else {
			finalCipherList = finalCipherList + "," + ciphersMapping[cipherId]
		}
	}

	/*
		Libcurl code to set the ssl ciphers to be used during connection.
		curl_easy_setopt(hnd, CURLOPT_SSL_CIPHER_LIST, "rsa_aes_128_sha,rsa_aes_256_sha");
	*/

	myCurl := this.myCurl
	myCurl.Setopt(curl.OPT_SSL_CIPHER_LIST, finalCipherList)
}

func setResponseSize(maxSize int64) (finalValue int64) {
	/*
			 get the first 200 bytes
			 curl_easy_setopt(curl, CURLOPT_RANGE, "0-199")

			 The unfortunate part is that for HTTP, CURLOPT_RANGE is not always enforced.
			 In this case we want to be able to still restrict the amount of data written
			 to the buffer.

			 For now we shall not use this. In the future, if the option becomes enforced
			 for HTTP then it can be used.

			 finalRange := "0-" + fmt.Sprintf("%s", MIN_REQUEST_SIZE)
		     finalRange = "0-" + fmt.Sprintf("%s", MAX_REQUEST_SIZE)
		     finalRange = "0-" + fmt.Sprintf("%s", maxSize)

		     myCurl := this.myCurl
		     myCurl.Setopt(curl.OPT_RANGE, finalRange)
	*/
	// Max Value = 64MB
	// Min Value = 20MB

	finalValue = MIN_RESPONSE_SIZE

	if maxSize > MAX_RESPONSE_SIZE {
		finalValue = MAX_RESPONSE_SIZE
	} else if (maxSize <= MAX_RESPONSE_SIZE) && (maxSize >= MIN_RESPONSE_SIZE) {
		finalValue = maxSize
	}

	return

}

func findPath() (string, error) {
	subpath := filepath.FromSlash(_PATH + "curl_whitelist.json")

	// Get directory of currently running file.
	listPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "", fmt.Errorf(" n1qlcerts directory does not exist under .../var/lib/couchbase/  on node " + hostname)
	}

	// Get the full path to curl_whitelist.json
	listPath = listPath + subpath
	return listPath, nil
}

func initialCheck() error {
	listPath, err := findPath()
	if err != nil {
		return err
	}

	// Open the file to see if it exists
	// 1. If it does not exist, then return with error. (Disable access to the CURL function)

	_, err = os.Stat(listPath)
	if os.IsNotExist(err) {
		// curl_whitelist.json not exist
		return fmt.Errorf("File ../Couchbase/var/lib/couchbase/n1qlcerts/curl_whitelist.json does not exist on node" + hostname + ". CURL() end points should be whitelisted.")
	}

	// Some other error other than not exists.
	return err

}

func whitelistCheck(url string) error {
	// 1. See initialCheck()..

	listPath, err := findPath()
	if err != nil {
		return err
	}

	// Read file
	b, err := ioutil.ReadFile(listPath)
	if err != nil {
		return err
	}

	// 2.a. If file curl_whitelist.json exists but is empty, then all access is false. But since all fields
	//    are treated as empty, it denies access to CURL. (same as above)
	if len(b) == 0 {
		return fmt.Errorf("File ../Couchbase/var/lib/couchbase/n1qlcerts/curl_whitelist.json is empty on node " + hostname + ". CURL() end points should be whitelisted.")
	}

	// 2.b. If it exists, is not empty but is invalid JSON (anything except JSON object), then NO ACCESS.
	var list map[string]interface{}

	if err := json.Unmarshal(b, &list); err != nil {
		return fmt.Errorf("File ../Couchbase/var/lib/couchbase/n1qlcerts/curl_whitelist.json contains invalid JSON on node " + hostname + ". Contents have to be a JSON object.")
	}

	// 2.c If it exists, and is valid json - Check entries.
	// 2.c.i. If all entries are invalid - Treat same as 2.
	//        Invalid entries are - {}, populated values not containing the field all_access.

	if len(list) == 0 {
		return fmt.Errorf("File ../Couchbase/var/lib/couchbase/n1qlcerts/curl_whitelist.json contains empty JSON object on node " + hostname + ". CURL() end points should be whitelisted.")
	}

	allaccess, ok := list["all_access"]
	if !ok {
		return fmt.Errorf("Boolean field all_access does not exist in file ../Couchbase/var/lib/couchbase/n1qlcerts/curl_whitelist.json on node " + hostname + ".")
	}

	// 2.c.ii. If all_access false - Use only those entries that are valid.

	// Structure is as follows
	// {
	//  "all_access":true/false,
	//  "allowed_urls":[ list of urls ],
	//  "disallowed_urls":[ list of urls ],
	// }
	isOk := checkType(allaccess, true)

	if !isOk {
		// Type check error
		return fmt.Errorf("all_access should be boolean value in file ../Couchbase/var/lib/couchbase/n1qlcerts/curl_whitelist.json on node " + hostname + ".")
	}

	if !allaccess.(bool) {

		notAccessable := false

		// Restricted access based on fields allowed_urls and disallowed_urls
		disallowedUrls, ok := list["disallowed_urls"]
		if ok {
			isOk := checkType(disallowedUrls, false)
			if !isOk {
				// Type check error
				return fmt.Errorf("disallowed_urls should be list of urls in file ../Couchbase/var/lib/couchbase/n1qlcerts/curl_whitelist.json on node " + hostname + ".")
			}
			// Valid values. Disallowed urls get 1st preference.
			disallow, err := sliceContains(disallowedUrls.([]interface{}), url)
			if err == nil && disallow {
				return fmt.Errorf("URL end point isn't whitelisted " + url + " on node " + hostname + ".")
			}
			if err != nil {
				return err
			}
		}

		allowedUrls, ok := list["allowed_urls"]
		if ok {
			isOk := checkType(allowedUrls, false)
			if !isOk {
				// Type check error
				return fmt.Errorf("allowed_urls should be list of urls in file ../Couchbase/var/lib/couchbase/n1qlcerts/curl_whitelist.json on node " + hostname + ".")
			}

			// If in allowed_urls then this query is valid.
			allow, err := sliceContains(allowedUrls.([]interface{}), url)
			if err == nil && allow {
				return nil
			} else {
				if err != nil {
					return err
				}
				// If it isnt in the allowed_urls
				notAccessable = true
			}
		} else {
			// allowed_urls is empty.
			notAccessable = true
		}

		// URL is not present in disallowed url and is not in allowed_urls.
		if notAccessable {
			// If it reaches here, then the url isnt in the allowed_urls or the prefix_urls, and is also
			// not in the disallowed urls.
			return fmt.Errorf("URL end point isn't whitelisted " + url + " on node " + hostname + ".")
		}

	}

	//  2.c.iii. If all_access true - FULL CURL ACCESS

	return nil
}

// Type assertion for whitelist fields
func checkType(val interface{}, accessfield bool) bool {
	if accessfield {
		_, ok := val.(bool)
		return ok
	} else {
		_, ok := val.([]interface{})
		return ok
	}

	return true
}

// Check if urls fields in whitelist contain the input url
func sliceContains(field []interface{}, url string) (bool, error) {
	for _, val := range field {
		nVal, ok := val.(string)
		if !ok {
			return false, fmt.Errorf("Both allowed_urls and disallowed urls should be list of url strings.")
		}
		// Check if list of values is a prefix of input url
		if strings.HasPrefix(url, nVal) {
			return true, nil
		}
	}
	return false, nil
}

/* Other auth values
var (
			CURLAUTH_NONE    = 0        /* nothing
			CURLAUTH_BASIC   = (1 << 0) /* Basic (default)
			CURLAUTH_DIGEST  = (1 << 1) /* Digest
			CURLAUTH_ANYSAFE = (^CURLAUTH_BASIC)
		)
*/
