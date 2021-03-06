package command

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/munisystem/rosculus/deployment"
)

type NewCommand struct {
	Meta

	name                       string
	sourceDBInstanceIdentifier string
	availabilityZone           string
	dbSubnetGroupName          string
	dbInstanceIdentifierBase   string
	dbInstanceTags             string
	dbMasterUserPassword       string
	publiclyAccessible         bool
	dbInstanceClass            string
	vpcSecurityGroupIdsString  string
	dnsimpleAuthToken          string
	dnsimpleAccountID          string
	dnsimpleDomain             string
	dnsimpleRecordID           int
	dnsimpleRecordName         string
	dnsimpleRecordTTL          int
	rollback                   bool
}

func (c *NewCommand) Run(args []string) int {
	if err := c.parseArgs(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	bucket := os.Getenv("AWS_S3_BUCKET_NAME")
	if bucket == "" {
		fmt.Fprintln(os.Stderr, errors.New("Please set s3 bucket name in AWS_S3_BUCKET_NAME"))
		return 1
	}

	var securityGroups []string
	if c.vpcSecurityGroupIdsString != "" {
		securityGroups = strings.Split(c.vpcSecurityGroupIdsString, ",")
	}

	instanceTags := make(map[string]string)
	if c.dbInstanceTags != "" {
		keyValues := strings.Split(c.dbInstanceTags, ",")
		for _, keyValue := range keyValues {
			arr := strings.Split(keyValue, "=")
			if len(arr) != 2 {
				fmt.Fprintln(os.Stderr, errors.New("-db-instance-tags is illegal format, please set like 'key1=value1,key2=value2'"))
				return 1
			}
			instanceTags[arr[0]] = arr[1]
		}
	}

	dep := &deployment.Deployment{
		SourceDBInstanceIdentifier: c.sourceDBInstanceIdentifier,
		DBMasterUserPassword:       c.dbMasterUserPassword,
		DBInstanceTags:             instanceTags,
		AvailabilityZone:           c.availabilityZone,
		DBSubnetGroupName:          c.dbSubnetGroupName,
		PubliclyAccessible:         c.publiclyAccessible,
		DBInstanceClass:            c.dbInstanceClass,
		VPCSecurityGroupIds:        securityGroups,
		DNSimple: deployment.DNSimple{
			AuthToken:  c.dnsimpleAuthToken,
			AccountID:  c.dnsimpleAccountID,
			Domain:     c.dnsimpleDomain,
			RecordID:   c.dnsimpleRecordID,
			RecordName: c.dnsimpleRecordName,
			TTL:        c.dnsimpleRecordTTL,
		},
		Current: deployment.Current{
			InstanceIdentifier: c.dbInstanceIdentifierBase + "-blue",
			Endpoint:           "",
		},
		Previous: deployment.Previous{
			InstanceIdentifier: c.dbInstanceIdentifierBase + "-green",
			Endpoint:           "",
		},
		Rollback: c.rollback,
	}

	if err := dep.Put(bucket, c.name); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func (c *NewCommand) parseArgs(args []string) error {
	flag := flag.NewFlagSet("rosculus", flag.ContinueOnError)

	flag.StringVar(&c.sourceDBInstanceIdentifier, "source-db-instance-identifier", "", "SourceDBInstanceIdentifier")
	flag.StringVar(&c.dbInstanceIdentifierBase, "db-instance-identifier-base", "", "DBInstanceIdentifierBase")
	flag.StringVar(&c.dbMasterUserPassword, "db-master-user-password", "", "DBMasterUserPassword")
	flag.StringVar(&c.dbInstanceTags, "db-instance-tags", "", "DBInstanceTags")
	flag.StringVar(&c.availabilityZone, "availability-zone", "", "AvailabilityZone")
	flag.StringVar(&c.dbSubnetGroupName, "db-subnet-group-name", "", "DBSubnetGroupName")
	flag.BoolVar(&c.publiclyAccessible, "publicly-accessible", true, "PubliclyAccessible")
	flag.StringVar(&c.dbInstanceClass, "db-instance-class", "db.m3.medium", "DBInstanceClass")
	flag.StringVar(&c.vpcSecurityGroupIdsString, "vpc-security-group-ids", "", "VPCSecurityGroupIds")
	flag.StringVar(&c.dnsimpleAuthToken, "dnsimple-auth-token", "", "DNSimpleAuthToken")
	flag.StringVar(&c.dnsimpleAccountID, "dnsimple-account-id", "", "DNSimpleAccountID")
	flag.StringVar(&c.dnsimpleDomain, "dnsimple-domain", "", "DNSimpleDomain")
	flag.IntVar(&c.dnsimpleRecordID, "dnsimple-record-id", 0, "DNSimpleRecordID")
	flag.StringVar(&c.dnsimpleRecordName, "dnsimple-record-name", "", "DNSimpleRecordName")
	flag.IntVar(&c.dnsimpleRecordTTL, "dnsimple-record-ttl", 60, "DNSimpleRecordTTL")
	flag.BoolVar(&c.publiclyAccessible, "rollback", true, "Rollback")

	if err := flag.Parse(args); err != nil {
		return err
	}

	if c.sourceDBInstanceIdentifier == "" {
		return errors.New("Please specify original DB instance identifier")
	}

	if c.dbInstanceIdentifierBase == "" {
		return errors.New("Please specify DB instance identifier base")
	}

	if c.availabilityZone == "" {
		return errors.New("Please specify DB instance AvailabilityZone")
	}

	if c.dbSubnetGroupName == "" {
		return errors.New("Please specify DB instance SubnetGroupName")
	}

	if 0 < flag.NArg() {
		c.name = flag.Arg(0)
	}

	if c.name == "" {
		return errors.New("Please specify deployment name")
	}

	return nil
}

func (c *NewCommand) Synopsis() string {
	return ""
}

func (c *NewCommand) Help() string {
	helpText := `

`
	return strings.TrimSpace(helpText)
}
