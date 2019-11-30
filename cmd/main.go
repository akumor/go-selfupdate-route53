// Updates AWS Route53 record with pulic IP
package main

import (
	"log"
	"net"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"
)

func main() {
	//region := ""
	//start := time.Now()

	// Define host port id outputfile flags
	/*apiHost := flag.String("host", "localhost", "API host to query for the wishlist.")
	apiPort := flag.String("port", "80", "API port to query for the wishlist.")
	id := flag.String("id", "DEFAULT", "ID of the Amazon wishlist.")

	flag.Parse()
	*/

	// Initial credentials loaded from SDK's default credential chain. Such as
	// the environment, shared credentials (~/.aws/credentials), or EC2 Instance
	// Role.
	/*
		var sess *session.Session
		if region == "" {
			sess = session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))
		} else {
			sess = session.Must(session.NewSession(&aws.Config{
				Region: aws.String(region)},
			))
		}*/

	//oldIP := ""
	for {

		// Attempt to get the public IP
		publicIP, err := getPublicIP()
		if err != nil {
			// TODO: allow sleep duration to be configureable
			log.Printf("getting public IP failed. will retry in %d seconds\n", 300)
			time.Sleep(time.Duration(300) * time.Second)
			continue
		}
		log.Printf("Obtained public IP: %s\n", publicIP)

		/*
			if oldIP != publicIP {
				// Create a Route53 client from just a session.
				r53 := route53.New(sess)
				createCNAME(r53)
				if err != nil {
					log.Println("failed to update AWS Route 53. will retry in %d seconds", 300)
					time.Sleep(300)
					continue
				}
				log.Println("Route53 record updated")
				oldIP = publicIP
			} else {
				log.Println("Public IP did not change. Nothing to do.")
			}
		*/

		// TODO: Allow sleep duration to be configurable
		time.Sleep(time.Duration(300) * time.Second)
	}

	//log.Printf("%.2fs elapsed\n", time.Since(start).Seconds())
}

func createCNAME(svc route53iface.Route53API, zoneID, name, target string, ttl int64) error {
	params := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{ // Required
			Changes: []*route53.Change{ // Required
				{ // Required
					Action: aws.String("UPSERT"), // Required
					ResourceRecordSet: &route53.ResourceRecordSet{ // Required
						Name: aws.String(name),    // Required
						Type: aws.String("CNAME"), // Required
						ResourceRecords: []*route53.ResourceRecord{
							{ // Required
								Value: aws.String(target), // Required
							},
						},
						TTL:           aws.Int64(ttl),
						SetIdentifier: aws.String("Arbitrary Id describing this change set"),
					},
				},
			},
			Comment: aws.String("Sample update."),
		},
		HostedZoneId: aws.String(zoneID), // Required
	}
	_, err := svc.ChangeResourceRecordSets(params)
	if err != nil {
		return err
	}
	// TODO: wait for Route53 DNS servers to become in sync: https://github.com/aws/aws-sdk-go/blob/v1.25.42/service/route53/api.go#L2489
	return nil
}

// getPublicIP returns the public ip address
func getPublicIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().String()
	idx := strings.LastIndex(localAddr, ":")
	return localAddr[0:idx], nil
}

// isPublicIP returns true if provided IP is a valid public IP and false otherwise
func isPublicIP(IP net.IP) bool {
	//Class       Starting IPAddress     Ending IP Address    # of Hosts
	//A           10.0.0.0               10.255.255.255       16,777,216
	//B           172.16.0.0             172.31.255.255       1,048,576
	//C           192.168.0.0            192.168.255.255      65,536
	//Link-local  169.254.0.0            169.254.255.255      65,536
	//Local       127.0.0.0              127.255.255.255      16777216
	if IP.IsLoopback() || IP.IsLinkLocalMulticast() || IP.IsLinkLocalUnicast() {
		return false
	}
	if ip4 := IP.To4(); ip4 != nil {
		switch {
		case ip4[0] == 10:
			return false
		case ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31:
			return false
		case ip4[0] == 192 && ip4[1] == 168:
			return false
		default:
			return true
		}
	}
	return false
}
