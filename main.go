package main

import (
    "time"
    "os"
    "os/signal"
    "log"
    "fmt"
    "strconv"
    "net"
    "golang.org/x/net/icmp"
    "golang.org/x/net/ipv4"
    "golang.org/x/net/ipv6"
)

// max returns the larger of x or y.
func max(x, y int64) int64 {
    if x < y {
        return y
    }
    return x
}

// min returns the smaller of x or y.
func min(x, y int64) int64 {
    if x > y {
        return y
    }
    return x
}

func SetTTLProtcolWise(c *icmp.PacketConn, protocol_choice string, ttl int){
        if protocol_choice == "IPv4"{
            c.IPv4PacketConn().SetTTL(ttl);
        } else if protocol_choice == "IPv6"{
            c.IPv6PacketConn().SetHopLimit(ttl);
        } else {
        }
}

func Ping(addr string, protocol_choice string, ttl int, maxTimeWaitForReply int) (*net.IPAddr, time.Duration, error) {

    // If IPV4 or IPV6, Listening Addr changes
    var listenerAddr string
    var network string
    var listen_packet_protocol string
    var parse_msg_protocol int

    if protocol_choice == "IPv4"{

        listenerAddr = "0.0.0.0"
        listen_packet_protocol = "ip4:icmp"
        network = "ip4"
        parse_msg_protocol = ipv4.ICMPTypeEcho.Protocol()
    } else if protocol_choice == "IPv6"{
        listenerAddr = "::"
        listen_packet_protocol = "ip6:ipv6-icmp"
        network = "ip6"
        parse_msg_protocol = ipv6.ICMPTypeEchoRequest.Protocol()
    } else {
    }
        //ListenPacket listens for incoming ICMP packets addressed to listenerAddr
        c, err := icmp.ListenPacket(listen_packet_protocol, listenerAddr)
        if err != nil {
            return nil, 0, err
        }

        if ttl != 0 {
            SetTTLProtcolWise(c,protocol_choice, ttl);
        }
        defer c.Close()

        // Get the real IP of the host/target
        destIP, err := net.ResolveIPAddr(network, addr)
        if err != nil {
            panic(err)
            return nil, 0, err
        }

        // Make a new ICMP message
        msg := icmp.Message{
            Code: 0,
            Body: &icmp.Echo{
                ID: os.Getpid() & 0xffff,
                Seq: 1,
                Data: []byte(""),
            },
        }

        if protocol_choice == "IPv4"{
            msg.Type  = ipv4.ICMPTypeEcho
        } else if protocol_choice == "IPv6"{
            msg.Type = ipv6.ICMPTypeEchoRequest
        } else {
           //TODO
        }

        bytes, err := msg.Marshal(nil)
        if err != nil {
            return destIP, 0, err
        }

        // Send ICMP Echo Request
        start := time.Now()
        numOfBytes, err := c.WriteTo(bytes, destIP)
        if err != nil {
            return destIP, 0, err
        } else if numOfBytes != len(bytes) {//If num of bytes written is not what's expected to send
            return destIP, 0, fmt.Errorf("Could Only Write  %v; Expected to write  %v", numOfBytes, len(bytes))
        }

        reply := make([]byte, 1500)

        //Wait for the duration and no further 
        err = c.SetReadDeadline(time.Now().Add(time.Duration(maxTimeWaitForReply) * time.Millisecond))
        if err != nil {
            return destIP, 0, err
        }


        //Read ICMP Reply
        numOfBytes, peer, err := c.ReadFrom(reply)
        if err != nil {
            return destIP, 0, err
        }

        //Calculate the latency 
        duration := time.Since(start)

       //Parse the Reply as per protocol
        replyMsg, err := icmp.ParseMessage(parse_msg_protocol, reply[:numOfBytes])
        if err != nil {
            return destIP, 0, err
        }

        switch replyMsg.Type {
            case ipv4.ICMPTypeEchoReply:
                return destIP, duration, nil
            case ipv6.ICMPTypeEchoReply:
                return destIP, duration, nil
            default:
                return destIP, 0, fmt.Errorf("got %+v from %v; Expected an echo reply", replyMsg, peer)
        }

}

func main() {

    if len(os.Args) != 5 {
        fmt.Printf("Usage is sudo ./main [hostname/Ip Address] [IPv4 or IPv6] [TTL Duration (milisec)] [maxTimeWaitForReply (milisec)]\n")
        return
    }

    hostname := os.Args[1]
    protocol_choice := os.Args[2]
    ttl,_ := strconv.Atoi(os.Args[3])
    maxTimeWaitForReply,_ := strconv.Atoi(os.Args[4])

    signalChan := make(chan os.Signal, 1)
    signal.Notify(signalChan, os.Interrupt)

    packetSent := int64(0)
    replyReceived := int64(0)
    duration := int64(0)
    max_duration := int64(0)
    min_duration := int64(1000) //large enough
    ticker := time.NewTicker(1 * time.Second)
    for {
        select{
            case <-ticker.C:
                destIP, dur, err := Ping(hostname,protocol_choice,ttl,maxTimeWaitForReply)
                packetSent += 1
                if err != nil {
                    log.Printf("Ping %s (%s): %s\n", hostname, destIP, err)
                } else {
                    log.Printf("Ping %s (%s): %s\n", hostname, destIP, dur)
                    replyReceived += 1
                    duration += dur.Milliseconds()
                    max_duration = max(dur.Milliseconds(),max_duration)
                    min_duration = min(dur.Milliseconds(),min_duration)
                }
            case <-signalChan:
               fmt.Printf("\n Packets Sent : %v\n", packetSent)
               fmt.Printf("\n Replies Received : %v\n",replyReceived)
               fmt.Printf("\n Loss : %v Percent\n",((packetSent-replyReceived)/packetSent * 100))
               if replyReceived > 0 {
                   fmt.Printf("\n Avg Latency : %v ms\n",(duration)/replyReceived)
                   fmt.Printf("\n Max Latency : %v ms\n",max_duration)
                   fmt.Printf("\n Min Latency : %v ms\n",min_duration)
               }
               return
            default:
        }
    }

}

