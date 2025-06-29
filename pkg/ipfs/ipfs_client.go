package ipfs

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// IPFSClient represents an IPFS client with its scheme, host, port and LassieClientURL details
type IPFSClient struct {
	// Scheme is the protocol scheme (e.g. http, https) used for communication with the IPFS node
	scheme string

	// Host is the hostname or IP address of the IPFS node
	host string

	// Port is the TCP port number used for communication with the IPFS node
	port int

	// LassieClientURL is the URL of the Lassie client
	LassieClientURL string
}

// NewIPFSClient creates a new instance of IPFSClient with the provided scheme, host, and port.
// It also sets the LassieClientURL based on the provided scheme, host, and port.
func NewIPFSClient(scheme string, host string, port int) *IPFSClient {

	return &IPFSClient{
		scheme:          scheme,
		host:            host,
		port:            port,
		LassieClientURL: fmt.Sprintf("%s://%s:%d/ipfs", scheme, host, port),
	}

}

// GetURLFromCID returns the URL of the given CID on the IPFS node.
//
// The function constructs the URL by appending the provided CID to the LassieClientURL of the IPFSClient.
//
// Parameters:
//   - cid (string) - The content identifier (CID) of the data to retrieve.
//
// Returns:
//   - cidUrl (string) - The constructed URL of the given CID.
//   - err (error) - An error if any occurs while constructing the URL.
func (ipfsC *IPFSClient) GetURLFromCID(cid string) (cidUrl string, err error) {
	cidUrl = fmt.Sprintf("%s/%s", ipfsC.LassieClientURL, cid)

	return cidUrl, err
}

// GetDataFromCID retrieves the data associated with the given CID from the IPFS node.
//
// The function constructs an HTTP GET request to the LassieClientURL of the IPFSClient with the provided CID appended.
// It then sends the request and reads the response body into a byte slice.
//
// Parameters:
//   - cid (string) - The content identifier (CID) of the data to retrieve.
//
// Returns:
//   - data ([]byte) - The retrieved data from the IPFS node.
//   - err (error) - An error if any occurs while constructing the request or reading the response.
func (ipfsC *IPFSClient) GetDataFromCID(cid string) (data []byte, err error) {
	cidUrl := fmt.Sprintf("%s/%s", ipfsC.LassieClientURL, cid)

	req, err := http.NewRequest("GET", cidUrl, nil)
	req.Header.Set("Accept", "*/*")
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: time.Second * 20}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return
}
