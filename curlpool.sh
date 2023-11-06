#!/bin/bash
 
function make_curl_request() {
    """
    This function makes a curl request to the specified endpoint with the given parameters.
    
    Parameters:
    endpoint (str): The URL endpoint to send the request to
    data (str): The data to be sent in the request body
    
    Returns:
    None
    """
    endpoint="http://localhost:1234/testfile.css"    
    # Make the curl request with the specified parameters
    curl -X 'GET' "$endpoint" 
}
 
# Set the maximum number of concurrent requests to 10
max_concurrent_requests=5
 
# Set the total number of requests to 50
total_requests=5
 
# Calculate the number of iterations needed to reach the total number of requests
num_iterations=$((total_requests / max_concurrent_requests))
 
# Loop through the iterations and make the curl requests
for ((i=0; i<$num_iterations; i++)); do
    # Loop through the maximum number of concurrent requests and make the curl requests
    for ((j=0; j<$max_concurrent_requests; j++)); do
        make_curl_request &
    done
    
    # Wait for all the requests to finish before continuing to the next iteration
    wait
done