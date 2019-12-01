// Updates AWS Route53 record with pulic IP
package main

import (
	"flag"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"

	externalip "github.com/akumor/go-external-ip"
)

func main() {
	// Define host port id outputfile flags
	hostedZoneID := flag.String("hosted-zone-id", "localhost", "AWS Route 53 Hosted Zone ID.")
	recordName := flag.String("record-name", "80", "DNS record name to upsert.")
	ttl := flag.Int64("ttl", 300, "TTL of the DNS record.")
	region := flag.String("region", "us-east-1", "AWS region to use for Route 53.")

	flag.Parse()

	// Initial credentials loaded from SDK's default credential chain. Such as
	// the environment, shared credentials (~/.aws/credentials), or EC2 Instance
	// Role.
	var sess *session.Session
	sess = session.Must(session.NewSession(&aws.Config{
		Region: aws.String(*region)},
	))
	// Create a Route53 client from just a session.
	r53 := route53.New(sess)

	oldIP := ""
	// Create the default consensus,
	// using the default configuration and no logger.
	consensus := externalip.DefaultConsensus(nil, nil)
	for {
		// Get the public IP,
		// which is never <nil> when err is <nil>.
		publicIP, err := consensus.ExternalIP()
		if err != nil {
			// TODO: allow sleep duration to be configureable
			log.Printf("getting public IP failed. will retry in %d seconds\n", 300)
			time.Sleep(time.Duration(300) * time.Second)
			continue
		}
		log.Printf("Obtained public IP: %s\n", publicIP.String()) // print IPv4/IPv6 in string format

		publicIPstr := publicIP.String()
		isIPv4 := false
		for i := 0; i < len(publicIPstr); i++ {
			switch publicIPstr[i] {
			case '.':
				isIPv4 = true
				break
			case ':':
				isIPv4 = false
				break
			}
		}

		if oldIP != publicIP.String() {
			recordType := ""
			if isIPv4 {
				log.Println("detected IPv4 IP address")
				recordType = "A"
			} else {
				log.Println("detected IPv6 IP address")
				recordType = "AAAA"
			}
			log.Printf("Creating %s record type in hosted zone %s with name %s and IP %s\n", recordType, *hostedZoneID, *recordName, publicIPstr)
			err = createRecord(r53, recordType, *hostedZoneID, *recordName, publicIPstr, *ttl)
			if err != nil {
				log.Println(err)
				log.Printf("failed to update AWS Route 53. will retry in %d seconds\n", 300)
				time.Sleep(time.Duration(300) * time.Second)
				continue
			}
			log.Println("Route53 record updated")
			oldIP = publicIP.String()
		} else {
			log.Println("Public IP did not change. Nothing to do.")
		}

		// TODO: Allow sleep duration to be configurable
		time.Sleep(time.Duration(300) * time.Second)
	}
}

// createRecord performs an UPSERT of a DNS record AWS Route 53
func createRecord(svc route53iface.Route53API, recordType, zoneID, name, target string, ttl int64) error {
	params := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{ // Required
			Changes: []*route53.Change{ // Required
				{ // Required
					Action: aws.String("UPSERT"), // Required
					ResourceRecordSet: &route53.ResourceRecordSet{ // Required
						Name: aws.String(name),       // Required
						Type: aws.String(recordType), // Required
						ResourceRecords: []*route53.ResourceRecord{
							{ // Required
								Value: aws.String(target), // Required
							},
						},
						TTL:           aws.Int64(ttl),
						Weight:        aws.Int64(100),
						SetIdentifier: aws.String("Arbitrary Id describing this change set"),
					},
				},
			},
			Comment: aws.String("Sample update."),
		},
		HostedZoneId: aws.String(zoneID), // Required
	}
	output, err := svc.ChangeResourceRecordSets(params)
	if err != nil {
		return err
	}
	// TODO: wait for Route53 DNS servers to become in sync: https://github.com/aws/aws-sdk-go/blob/v1.25.42/service/route53/api.go#L2489
	log.Printf(output.String())
	return nil
}
