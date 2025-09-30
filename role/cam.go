package role

import (
	"context"
	"time"

	"cucats.org/discord/config"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"golang.org/x/oauth2"
)

const (
	tenantID  = "49a50445-bdfa-4b79-ade3-547b4f3986e9"
	authority = "https://login.microsoftonline.com/" + tenantID
)

var CamOAuth *oauth2.Config

func InitCamOAuth() {
	CamOAuth = &oauth2.Config{
		ClientID:     config.CamClientID,
		ClientSecret: config.CamClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authority + "/oauth2/v2.0/authorize",
			TokenURL: authority + "/oauth2/v2.0/token",
		},
		RedirectURL: config.Host + "/cam/callback",
		Scopes:      []string{"User.Read"},
	}
}

type TokenCredential struct {
	accessToken string
}

func (tc *TokenCredential) GetToken(ctx context.Context, options policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{
		Token:     tc.accessToken,
		ExpiresOn: time.Now().Add(time.Hour),
	}, nil
}

type UserInfo struct {
	UPN         string
	DisplayName string
	IsStudent   bool
	IsAlumni    bool
	IsStaff     bool
	College     College
}

type College int

const (
	Unknown College = iota
	Christs
	Churchill
	Clare
	ClareHall
	CorpusChristi
	Darwin
	Downing
	Emmanuel
	Fitzwilliam
	Girton
	GonvilleCaius
	Homerton
	HughesHall
	Jesus
	Kings
	LucyCavendish
	Magdalene
	MurrayEdwards
	Newnham
	Pembroke
	Peterhouse
	Queens
	Robinson
	Selwyn
	SidneySussex
	StCatharines
	StEdmunds
	StJohns
	Trinity
	TrinityHall
	Wolfson
)

// Discord role IDs for colleges
var CollegeRoles = map[College]string{
	Christs:       "785993896642740294",
	Churchill:     "785993890578300928",
	Clare:         "785993904654516315",
	ClareHall:     "813071045082218526",
	CorpusChristi: "785993906668044319",
	Darwin:        "813071257784287323",
	Downing:       "785993908651950111",
	Emmanuel:      "785993911025008640",
	Fitzwilliam:   "785993912896585728",
	Girton:        "785993914418069524",
	GonvilleCaius: "785993916414951434",
	Homerton:      "785993918466097203",
	HughesHall:    "785993920098074665",
	Jesus:         "785993922249228329",
	Kings:         "785993924438523965",
	LucyCavendish: "785993925990809651",
	Magdalene:     "785993938615795723",
	MurrayEdwards: "785993941899804703",
	Newnham:       "785993946093322274",
	Pembroke:      "785993952418463765",
	Peterhouse:    "785993949260808263",
	Queens:        "785993955757391872",
	Robinson:      "785993958885949460",
	Selwyn:        "785993975780605952",
	SidneySussex:  "785993982533435402",
	StCatharines:  "785993985696989204",
	StEdmunds:     "785993989596774440",
	StJohns:       "785993994853023756",
	Trinity:       "785993997347979315",
	TrinityHall:   "785994000376135730",
	Wolfson:       "788466975726239794",
}

var collegeGroups = map[string]College{
	"ce5816d8-91da-44d5-b4c9-72291314b827": Christs,
	"1b270804-ec85-4171-8718-edd0dcda5f7d": Churchill,
	"d8ada922-61d3-4632-8707-5185b3bf10af": Clare,
	"74fa1c92-6c4b-492a-9316-32ea7dd4b65a": ClareHall,
	"27cdda96-1f23-44f2-8309-d081cf247af6": CorpusChristi,
	"9d2b6811-3455-4f50-8c6c-bcffaab6cc34": Darwin,
	"670f426c-0254-47a3-9a66-a08aa678dd16": Downing,
	"4061829b-d7f9-4f60-adb1-e7cda4b2d445": Emmanuel,
	"dd2d37a3-9a37-47df-ba3b-2a0dbcb8694f": Fitzwilliam,
	"7e5d91fc-4dee-4f6d-bd89-183d17c874d9": Girton,
	"b9f1cc7f-9a95-4075-82dc-3094ce09fe0e": GonvilleCaius,
	"a53ddffd-3d7b-447a-b545-59301cbeb8fd": Homerton,
	"8a5c602e-a5de-4a6a-8370-8a47ed9b329f": HughesHall,
	"f6a6bd78-7e78-46c2-bb9a-eec0621e43b1": Jesus,
	"cdb13798-6a29-4843-8539-1643bb9f5adb": Kings,
	"588ca32d-b1b3-4730-9b98-f1e2a6f17e18": LucyCavendish,
	"a8d8c12b-348f-46e7-8876-a4c2954ecf8b": Magdalene,
	"5cf716e5-71ff-49e1-9077-2c6cdc0f694b": MurrayEdwards,
	"29b9532a-a9fb-4ff3-ace4-5a0e54a60904": Newnham,
	"abe28c76-a6ac-46fa-9d31-b4c20d8b9624": Pembroke,
	"f8c3ec4e-5636-4b0f-9f89-1e21833949ae": Peterhouse,
	"31d3584f-aa76-4c4a-b880-b7f8b1e0107b": Queens,
	"df2d7e4a-caf6-444c-a554-b91df8e726b4": Robinson,
	"cf9e2f37-7fc4-4645-b109-55c5ed923833": Selwyn,
	"f9a44cdf-5ae9-49fe-93a4-73bf93427d64": SidneySussex,
	"65715637-1d4e-437f-aa14-b77850298db1": StCatharines,
	"39797ec4-08ae-4b94-aa1d-20368093d562": StEdmunds,
	"78011bf7-a974-4161-95e8-4c5acece8f6b": StJohns,
	"8ae6faff-1731-49e1-b786-6d8e08f9373a": Trinity,
	"3877b22b-b894-4a24-bfe5-17ef79ac90be": TrinityHall,
	"64bbfe87-dd43-4828-a035-731c576c0b3a": Wolfson,
}

const (
	guidStudent = "0cbcd7fb-1f17-48fc-ac3e-4a22131fa92d"
	guidStaff   = "1f440b90-597d-45b4-9a0d-11707f784de7"
	guidAlumni  = "bc7a045e-6775-423a-abc6-deac53b50712"
)

func GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	client, err := msgraphsdkgo.NewGraphServiceClientWithCredentials(&TokenCredential{accessToken: accessToken}, []string{})
	if err != nil {
		return nil, err
	}

	user, err := client.Me().Get(ctx, nil)
	if err != nil {
		return nil, err
	}

	userInfo := &UserInfo{
		UPN:         *user.GetUserPrincipalName(),
		DisplayName: *user.GetDisplayName(),
	}

	groups, err := client.Me().MemberOf().Get(ctx, nil)
	if err != nil {
		return nil, err
	}

	for _, group := range groups.GetValue() {
		if grp, ok := group.(*models.Group); ok {
			if guid := grp.GetId(); guid != nil {
				switch *guid {
				case guidStudent:
					userInfo.IsStudent = true
				case guidStaff:
					userInfo.IsStaff = true
				case guidAlumni:
					userInfo.IsAlumni = true
				default:
					if college, ok := collegeGroups[*guid]; ok {
						userInfo.College = college
					}
				}
			}
		}
	}

	return userInfo, nil
}
