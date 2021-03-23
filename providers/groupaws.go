package providers

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"net/url"
	"strings"
)

type Statement struct {
	Sid      string
	Effect   string
	Action   []string
	Resource []string
}

type PolicyDocument struct {
	Version   string
	Statement []Statement
}

func GetGroupID() [][2]string {

	// This array will contain objects having name and group id
	var groups_ids [][2]string

	// Get AWS session
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(
			"----------",
			"-------------",
			""),
	})
	if err != nil {
		fmt.Println("Error creating the session", err)
		//return
	}

	svc := iam.New(sess)

	// List IAM groups (*_Users and *_Admin)
	input := &iam.ListGroupsInput{}
	groups, err := svc.ListGroups(input)
	if err != nil {
		fmt.Println("Error listing groups", err)
		//return
	}

	for _, gr := range groups.Groups {

		// List groups inline policies
		input := &iam.ListGroupPoliciesInput{GroupName: aws.String(*gr.GroupName)}
		group_policies, err := svc.ListGroupPolicies(input)
		if err != nil {
			fmt.Println("Error listing group policies", err)
			//return
		}

		for _, policy_name := range group_policies.PolicyNames {

			// The policies that have the group ID's are named: policygen-*_Users-*
			if strings.Contains(*policy_name, "policygen") && strings.Contains(*policy_name, "Users") {

				// Retrieve policy information (the JSON data that is seen on the UI)
				input := &iam.GetGroupPolicyInput{
					GroupName:  aws.String(*gr.GroupName),
					PolicyName: aws.String(*policy_name),
				}
				response, err := svc.GetGroupPolicy(input)
				if err != nil {
					fmt.Println("Error GetGroupPolicy", err)
					//return
				}

				// Get the resource arn that contains the group ID from the policy JSON (requires decoding and unmarshal)
				decodedValue, err := url.QueryUnescape(*response.PolicyDocument) // this is a stringified json
				if err != nil {
					fmt.Println("Error decoding PolicyDocument", err)
					//return
				}
				var polDoc PolicyDocument
				json.Unmarshal([]byte(decodedValue), &polDoc)
				inline_policy_arn := polDoc.Statement[0].Resource[0]

				// Take the group ID from the arn
				inline_policy_arn = strings.Replace(inline_policy_arn, "arn:aws:iam::", "", -1)
				inline_policy_arn = strings.Replace(inline_policy_arn, ":role/AccountPowerUser", "", -1)

				// Add to the groups_ids array an object containing name and groupd ID
				var item [2]string
				item[0] = *gr.GroupName
				item[1] = inline_policy_arn
				groups_ids = append(groups_ids, item)

			}
		}
	}

	return groups_ids
}

func GetName() string {
	var name string
	var groups = GetGroupID()
	for i := 0; i < len(groups); i++ {
		name = (groups[i])[0]
	}
	return name
}
func GetID() string {
	var id string
	var groups = GetGroupID()
	for i := 0; i < len(groups); i++ {
		id = (groups[i])[1]
	}
	return id

}

/*func MatchNametoID( idFile string ) string{


	if idFile==GetID() {
		return name
	}
	return "name"
	//[[ALICE_Users 118709236800] [ATLAS_Users 124939657953] [BE-OP_TE-ABT_Users 011747925465] [CMS_Users 441541562849] [EP-CMG_Users 016245035593] [IT-DB-IA_Users 239533842733] [IT-DB-SAS_Users 610129916007] [LHCb_Users 744673412807] [Openlab_Users 639504270038] [TH_Users 654847038767]]
	//GETNAMe
}*/
