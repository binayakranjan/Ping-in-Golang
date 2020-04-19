
 Golang Vesion Used:
  go version go1.14.2 linux/amd64

 Additional Imports used :
   golang.org/x/net

 Execution Details:
   Build :  go build main.go
   Usage :  sudo ./main [hostname/Ip Address] [IPv4 or IPv6] [TTL Duration (milisec)] [maxTimeWaitForReply (milisec)]
   
   Description : The cli Tool expects 4 positional arguments such as ,
                 1. hostname or Ip Address of the destination
                 2. Whether IPv4 or IPv6 protocol used
                 3. TTL Duration , If TTL = 0, it behaves as if NO TTL is set
                 4. Maximum time duration ping should wait for reply

  Assumed Parameters, which could have been taken as user input:
          1. Packet size : 1500 bytes 
