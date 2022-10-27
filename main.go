package main

import "fmt"
import bencode "github.com/IncSW/go-bencode"
import "net"
import "bufio"
import "log"
import "net/http"
import "github.com/gorilla/mux"
import "encoding/json"
import "strings"
import "regexp"
import "io/ioutil"
import "math/rand"

type AllocationRequest struct {
    Ip      string `json:"ip"`
    Port    string `json:"port"`
    CallId  string `json:"callid"`
}

type AllocationResponse struct {
    Ip      string `json:"ip"`
    Port    string `json:"port"`
}

func allocate_offer(w http.ResponseWriter, r *http.Request) {
	fmt.Println("allocate offer-----------------")

	reqBody, _ := ioutil.ReadAll(r.Body)
	fmt.Println("Body:", reqBody)

	var allocationRequest AllocationRequest
	json.Unmarshal(reqBody, &allocationRequest)

	source_ip := allocationRequest.Ip
	source_port := allocationRequest.Port
	call_id := allocationRequest.CallId
	fmt.Printf("Allocate request, received IP: %s, port: %s (%s)\n", source_ip, source_port, call_id)

	var allocationResponse AllocationResponse
	allocationResponse = send_offer(allocationRequest)

	json.NewEncoder(w).Encode(allocationResponse)
}

func allocate_answer(w http.ResponseWriter, r *http.Request) {
	fmt.Println("allocate answer-----------------")

	reqBody, _ := ioutil.ReadAll(r.Body)
	fmt.Println("Body:", reqBody)

	var allocationRequest AllocationRequest
	json.Unmarshal(reqBody, &allocationRequest)

	source_ip := allocationRequest.Ip
	source_port := allocationRequest.Port
	call_id := allocationRequest.CallId
	fmt.Printf("Allocate request, received IP: %s, port: %s (%s)\n", source_ip, source_port, call_id)

	var allocationResponse AllocationResponse
	allocationResponse = send_answer(allocationRequest)

	json.NewEncoder(w).Encode(allocationResponse)
}

//	var dict interface{} = map[string]interface{} {
//		"command": "ping",
//	}

func send_offer(allocationRequest AllocationRequest) AllocationResponse {

//		cmd = {'command': 'offer',
//               'call-id': callid,
//               'from-tag': fromtag,
//               'sdp': sdp
//               }

	sdp := fmt.Sprintf("v=0\r\no=gv 2890844526 2890842807 IN IP4 10.47.16.5\r\ns= \r\nc=IN IP4 %s\r\nt=2873397496 2873404696\r\nm=audio %s RTP/AVP 0\r\n", allocationRequest.Ip, allocationRequest.Port)

	var request_data interface{} = map[string]interface{} {
		"command": "offer",
		"call-id": allocationRequest.CallId,
		"from-tag": "456",
		"sdp": sdp,
		"direction": []interface{}{"internal", "external"},
	}

	return send_request(request_data)
}

func send_answer(allocationRequest AllocationRequest) AllocationResponse {

//		cmd = {'command': 'answer',
//               'call-id': callid,
//               'from-tag': fromtag,
//               'to-tag': totag,
//               'sdp': sdp
//               }

	sdp := fmt.Sprintf("v=0\r\no=gv 2890844526 2890842807 IN IP4 10.47.16.5\r\ns= \r\nc=IN IP4 %s\r\nt=2873397496 2873404696\r\nm=audio %s RTP/AVP 0\r\n", allocationRequest.Ip, allocationRequest.Port)

	var request_data interface{} = map[string]interface{} {
		"command": "answer",
		"call-id": allocationRequest.CallId,
		"from-tag": "456",
		"to-tag": "789",
		"sdp": sdp,
	}

	return send_request(request_data)
}

func send_request(dict interface{}) AllocationResponse {
	fmt.Println("send_request -------------------------")

	data, err := bencode.Marshal(dict)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
	fmt.Println("-----Request prepared, sending------------")

	// Prepare a cookie
	random_n := rand.Intn(90000) + 10000
	cookie := fmt.Sprintf("cookie%d", random_n)

	// Format request
	request := fmt.Sprintf("%s %s", cookie, string(data))

	// TODO: get socket from a list
	connection, err := net.Dial("udp", "127.0.0.1:2223")
	if err != nil {
		fmt.Printf("Some error %v", err)
		return AllocationResponse{}
	}

	// Send request via connection
	fmt.Fprintf(connection, request)

	// Read response
	const maxBufferSize = 1024
	response := make([]byte, maxBufferSize)
	_, err = bufio.NewReader(connection).Read(response)

	if err != nil {
		fmt.Printf("ERROR writing to connection <======================= %s\n", err)
		return AllocationResponse{}
	}

	connection.Close()

	return parse_response(response)
}

func parse_response(response []byte) AllocationResponse {

	// Extract IP and port from RTPEngine's response
	//cookie001 d3:sdp175:v=0...

	fmt.Println("----------Decode response and return result---------")
	// Strip the cookie from the response and keep only the bencode
	words := strings.Fields(string(response))
	bencode_response := response[len(words[0])+1:]
	fmt.Printf("Bencode from response:%s\n\n", bencode_response)

	// Decode the bencode from data
	final_response, err := bencode.Unmarshal([]byte(bencode_response))
	if err != nil {
		fmt.Println("ERROR reading response <=============")
		return AllocationResponse{}
	}
	fmt.Println(final_response)

	final_response_map, _ := final_response.(map[string]interface{})

	result := string(final_response_map["result"].([]byte))

	if (result == "error") {
		fmt.Printf("=========> ERROR: %s\n", string(final_response_map["error-reason"].([]byte)))
		return AllocationResponse{}
	}

	final_response_string := string(final_response_map["sdp"].([]byte))
	fmt.Println("------------> Response SDP:", final_response_string)

	re := regexp.MustCompile("c=IN IP4 (.*?)\\r\\n")
	match := re.FindStringSubmatch(final_response_string)
	allocated_ip := match[1]

	fmt.Println("Allocate IP address:", allocated_ip)

	re = regexp.MustCompile("m=audio (.*?) RTP\\/AVP")
	match = re.FindStringSubmatch(final_response_string)
	allocated_port := match[1]

	fmt.Println("Allocated port:", allocated_port)

	fmt.Println("\n\nFinished with allocation.")

	allocationResponse := AllocationResponse{Ip: allocated_ip, Port: allocated_port}
	return allocationResponse
}

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/allocate_offer", allocate_offer).Methods("POST")
	myRouter.HandleFunc("/allocate_answer", allocate_answer).Methods("POST")

	log.Fatal(http.ListenAndServe(":10000", myRouter))
}

func main() {
	fmt.Println("Starting...")
	handleRequests()
}
