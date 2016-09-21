/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at
  http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package main

import (
	"encoding/json"
	"fmt"
    "strconv"
    "errors"
    "strings"
    "crypto/x509"
    "encoding/pem"
    "net/http"
    "net/url"
    "io/ioutil"
    "github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

func (t *SimpleChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Printf("Init called, initializing chaincode")
	
	var A, B string    // Entities
	var Aval, Bval int // Asset holdings
	var err error

	if len(args) != 4 {
		return nil, errors.New("Incorrect number of arguments. Expecting 4")
	}

	// Initialize the chaincode
	A = args[0]
	Aval, err = strconv.Atoi(args[1])
	if err != nil {
		return nil, errors.New("Expecting integer value for asset holding")
	}
	B = args[2]
	Bval, err = strconv.Atoi(args[3])
	if err != nil {
		return nil, errors.New("Expecting integer value for asset holding")
	}
	fmt.Printf("Aval = %d, Bval = %d\n", Aval, Bval)

	// Write the state to the ledger
	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return nil, err
	}

	err = stub.PutState(B, []byte(strconv.Itoa(Bval)))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// Transaction makes payment of X units from A to B
func (t *SimpleChaincode) invoke(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	fmt.Printf("Running invoke")
	
	var A, B string    // Entities
	var Aval, Bval int // Asset holdings
	var X int          // Transaction value
	var err error

	if len(args) != 3 {
		return nil, errors.New("Incorrect number of arguments. Expecting 3")
	}

	A = args[0]
	B = args[1]

	// Get the state from the ledger
	// TODO: will be nice to have a GetAllState call to ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		return nil, errors.New("Failed to get state")
	}
	if Avalbytes == nil {
		return nil, errors.New("Entity not found")
	}
	Aval, _ = strconv.Atoi(string(Avalbytes))

	Bvalbytes, err := stub.GetState(B)
	if err != nil {
		return nil, errors.New("Failed to get state")
	}
	if Bvalbytes == nil {
		return nil, errors.New("Entity not found")
	}
	Bval, _ = strconv.Atoi(string(Bvalbytes))

	// Perform the execution
	X, err = strconv.Atoi(args[2])
	Aval = Aval - X
	Bval = Bval + X
	fmt.Printf("Aval = %d, Bval = %d\n", Aval, Bval)

	// Write the state back to the ledger
	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return nil, err
	}

	err = stub.PutState(B, []byte(strconv.Itoa(Bval)))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// Deletes an entity from state
func (t *SimpleChaincode) delete(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	fmt.Printf("Running delete")
	
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 3")
	}

	A := args[0]

	// Delete the key from the state in ledger
	err := stub.DelState(A)
	if err != nil {
		return nil, errors.New("Failed to delete state")
	}

	return nil, nil
}

// Invoke callback representing the invocation of a chaincode
// This chaincode will manage two accounts A and B and will transfer X units from A to B upon invoke
func (t *SimpleChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Printf("Invoke called, determining function")
	
	// Handle different functions
	if function == "invoke" {
		// Transaction makes payment of X units from A to B
		fmt.Printf("Function is invoke")
		return t.invoke(stub, args)
	} else if function == "init" {
		fmt.Printf("Function is init")
		return t.Init(stub, function, args)
	} else if function == "delete" {
		// Deletes an entity from its state
		fmt.Printf("Function is delete")
		return t.delete(stub, args)
	}else if function == "varunWrite" {
		// Deletes an entity from its state
		fmt.Printf("Function is varunWrite")
		return t.varunWrite(stub, args)
	}else if function == "callerData" {
		// Deletes an entity from its state
		fmt.Printf("Function is callerData")
		return t.GetCallerdata(stub)
	}

	return nil, errors.New("Received unknown function invocation")
}

func (t* SimpleChaincode) Run(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Printf("Run called, passing through to Invoke (same function)")
	
	// Handle different functions
	if function == "invoke" {
		// Transaction makes payment of X units from A to B
		fmt.Printf("Function is invoke")
		return t.invoke(stub, args)
	} else if function == "init" {
		fmt.Printf("Function is init")
		return t.Init(stub, function, args)
	} else if function == "delete" {
		// Deletes an entity from its state
		fmt.Printf("Function is delete")
		return t.delete(stub, args)
	}

	return nil, errors.New("Received unknown function invocation")
}

// Query callback representing the query of a chaincode
func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Printf("Query called, determining function")
	
	if function == "query" {
		fmt.Printf("Function is query")
		return nil, errors.New("Invalid query function name. Expecting \"query\"")
	
	var A string // Entities
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the person to query")
	}

	A = args[0]

	// Get the state from the ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + A + "\"}"
		return nil, errors.New(jsonResp)
	}

	if Avalbytes == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + A + "\"}"
		return nil, errors.New(jsonResp)
	}

	jsonResp := "{\"Name\":\"" + A + "\",\"Amount\":\"" + string(Avalbytes) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)
	return Avalbytes, nil
	}else{
		return varunRead(stub, args)
	}
}


type cust struct {
	ID string `json:"id"`
	name string `json:"name"`
	age int `json:"age"`
}


func (t *SimpleChaincode) varunWrite(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	fmt.Println("ENTERING ********************** varunWrite ")

	c := cust{"24", "Varun Ojha", 32}
	bytes, err := json.Marshal(&c)

	err = stub.PutState("varun_name", bytes)
	if err != nil {
		return nil, err
	}
	
	return nil, nil
}

func varunRead(stub *shim.ChaincodeStub, args []string) ([]byte, error) {

	var str = args[0]
	if(str == "state"){
		bytes, err := stub.GetState("varun_name")
		return bytes, err
	}else{
		return []byte("Finally it works"), nil
	}
	
}


type ECertResponse struct {
    OK string `json:"OK"`
}   

//==============================================================================================================================
//   get_ecert - Takes the name passed and calls out to the REST API for HyperLedger to retrieve the ecert
//               for that user. Returns the ecert as retrived including html encoding.
//==============================================================================================================================
func (t *SimpleChaincode) GetEcert(stub *shim.ChaincodeStub, name string) ([]byte, error) {
    
    var cert ECertResponse
    
    peer_address, err := stub.GetState("Peer_Address")
                                                            if err != nil { return nil, errors.New("Error retrieving peer address") }

    response, err := http.Get("http://"+string(peer_address)+"/registrar/"+name+"/ecert")   // Calls out to the HyperLedger REST API to get the ecert of the user with that name
    
                                                            if err != nil { return nil, errors.New("Error calling ecert API") }
    
    defer response.Body.Close()
    contents, err := ioutil.ReadAll(response.Body)                  // Read the response from the http callout into the variable contents
    
                                                            if err != nil { return nil, errors.New("Could not read body") }
    
    err = json.Unmarshal(contents, &cert)
    
                                                            if err != nil { return nil, errors.New("Could not retrieve ecert for user: "+name) }
                                                            
    return []byte(string(cert.OK)), nil
}

//==============================================================================================================================
//   get_caller - Retrieves the username of the user who invoked the chaincode.
//                Returns the username as a string.
//==============================================================================================================================


func (t *SimpleChaincode) GetUsername(stub *shim.ChaincodeStub) (string, error) {

    bytes, err := stub.GetCallerCertificate();
                                                            if err != nil { return "", errors.New("Couldn't retrieve caller certificate") }
    x509Cert, err := x509.ParseCertificate(bytes);              // Extract Certificate from result of GetCallerCertificate                      
                                                            if err != nil { return "", errors.New("Couldn't parse certificate") }
                                                            
    return x509Cert.Subject.CommonName, nil
}

//==============================================================================================================================
//   check_affiliation - Takes an ecert as a string, decodes it to remove html encoding then parses it and checks the
//                      certificates common name. The affiliation is stored as part of the common name.
//==============================================================================================================================

func (t *SimpleChaincode) CheckAffiliation(stub *shim.ChaincodeStub, cert string) (int, error) {                                                                                                                                                                                                                   
    
    decodedCert, err := url.QueryUnescape(cert);                    // make % etc normal //
    
                                                            if err != nil { return -1, errors.New("Could not decode certificate") }
    
    pem, _ := pem.Decode([]byte(decodedCert))                           // Make Plain text   //

    x509Cert, err := x509.ParseCertificate(pem.Bytes);              // Extract Certificate from argument //
                                                        
                                                            if err != nil { return -1, errors.New("Couldn't parse certificate") }

    cn := x509Cert.Subject.CommonName
    
    res := strings.Split(cn,"\\")
    
    affiliation, _ := strconv.Atoi(res[2])
    
    return affiliation, nil
}

//==============================================================================================================================
//   get_caller_data - Calls the get_ecert and check_role functions and returns the ecert and role for the
//                   name passed.
//==============================================================================================================================

func (t *SimpleChaincode) GetCallerdata(stub *shim.ChaincodeStub) ([]byte, error){

    fmt.Println("Entering GetCallerdata")

    user, err := t.GetUsername(stub)
        if err != nil {
            fmt.Println("COULD NOT GET USERNAME %s", err); 
            return nil, err 
        }

        fmt.Println("USER: ")
        fmt.Println(user)

                                                                        
    ecert, err := t.GetEcert(stub, user);                   
        if err != nil {
            fmt.Println("COULD NOT GET ECERT %s", err); 
            return nil,  err 
        }

    fmt.Println("ecert: ")
    fmt.Println(ecert)

    affiliation, err := t.CheckAffiliation(stub,string(ecert));         
        if err != nil {
            fmt.Println("COULD NOT GET Affiliation %s", err); 
            return nil,  err 
        }

    fmt.Println("affiliation: ")
    fmt.Println(affiliation)

    return []byte(user), nil
}


func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}