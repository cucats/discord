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

const (
	UocUsersStaff   string = "1f440b90-597d-45b4-9a0d-11707f784de7"
	UocUsersStudent string = "0cbcd7fb-1f17-48fc-ac3e-4a22131fa92d"
	UocUsersAlumni  string = "bc7a045e-6775-423a-abc6-deac53b50712"
	UocUsersCamUPN  string = "b7a0f932-5964-41b2-9bb0-9b8cadf6b999"
	UocUsersGuests  string = "20c3c1f1-309f-497d-9169-3ac4907098a1"
	UocUsersAll     string = "cc2cdd8b-eace-4a4b-a950-9b989a183b97"

	ChristsMembers       string = "ce5816d8-91da-44d5-b4c9-72291314b827"
	ChurchillMembers     string = "1b270804-ec85-4171-8718-edd0dcda5f7d"
	ClareMembers         string = "d8ada922-61d3-4632-8707-5185b3bf10af"
	ClareHallMembers     string = "74fa1c92-6c4b-492a-9316-32ea7dd4b65a"
	CorpusChristiMembers string = "27cdda96-1f23-44f2-8309-d081cf247af6"
	DarwinMembers        string = "9d2b6811-3455-4f50-8c6c-bcffaab6cc34"
	DowningMembers       string = "670f426c-0254-47a3-9a66-a08aa678dd16"
	EmmanuelMembers      string = "4061829b-d7f9-4f60-adb1-e7cda4b2d445"
	FitzwilliamMembers   string = "dd2d37a3-9a37-47df-ba3b-2a0dbcb8694f"
	GirtonMembers        string = "7e5d91fc-4dee-4f6d-bd89-183d17c874d9"
	GonvilleCaiusMembers string = "b9f1cc7f-9a95-4075-82dc-3094ce09fe0e"
	HomertonMembers      string = "a53ddffd-3d7b-447a-b545-59301cbeb8fd"
	HughesHallMembers    string = "8a5c602e-a5de-4a6a-8370-8a47ed9b329f"
	JesusMembers         string = "f6a6bd78-7e78-46c2-bb9a-eec0621e43b1"
	KingsMembers         string = "cdb13798-6a29-4843-8539-1643bb9f5adb"
	LucyCavendishMembers string = "588ca32d-b1b3-4730-9b98-f1e2a6f17e18"
	MagdaleneMembers     string = "a8d8c12b-348f-46e7-8876-a4c2954ecf8b"
	MurrayEdwardsMembers string = "5cf716e5-71ff-49e1-9077-2c6cdc0f694b"
	NewnhamMembers       string = "29b9532a-a9fb-4ff3-ace4-5a0e54a60904"
	PembrokeMembers      string = "abe28c76-a6ac-46fa-9d31-b4c20d8b9624"
	PeterhouseMembers    string = "f8c3ec4e-5636-4b0f-9f89-1e21833949ae"
	QueensMembers        string = "31d3584f-aa76-4c4a-b880-b7f8b1e0107b"
	RobinsonMembers      string = "df2d7e4a-caf6-444c-a554-b91df8e726b4"
	SelwynMembers        string = "cf9e2f37-7fc4-4645-b109-55c5ed923833"
	SidneySussexMembers  string = "f9a44cdf-5ae9-49fe-93a4-73bf93427d64"
	StCatharinesMembers  string = "65715637-1d4e-437f-aa14-b77850298db1"
	StEdmundsMembers     string = "39797ec4-08ae-4b94-aa1d-20368093d562"
	StJohnsMembers       string = "78011bf7-a974-4161-95e8-4c5acece8f6b"
	TrinityMembers       string = "8ae6faff-1731-49e1-b786-6d8e08f9373a"
	TrinityHallMembers   string = "3877b22b-b894-4a24-bfe5-17ef79ac90be"
	WolfsonMembers       string = "64bbfe87-dd43-4828-a035-731c576c0b3a"
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
				case UocUsersStudent:
					userInfo.IsStudent = true
				case UocUsersStaff:
					userInfo.IsStaff = true
				case UocUsersAlumni:
					userInfo.IsAlumni = true
				case ChristsMembers:
					userInfo.College = Christs
				case ChurchillMembers:
					userInfo.College = Churchill
				case ClareMembers:
					userInfo.College = Clare
				case ClareHallMembers:
					userInfo.College = ClareHall
				case CorpusChristiMembers:
					userInfo.College = CorpusChristi
				case DarwinMembers:
					userInfo.College = Darwin
				case DowningMembers:
					userInfo.College = Downing
				case EmmanuelMembers:
					userInfo.College = Emmanuel
				case FitzwilliamMembers:
					userInfo.College = Fitzwilliam
				case GirtonMembers:
					userInfo.College = Girton
				case GonvilleCaiusMembers:
					userInfo.College = GonvilleCaius
				case HomertonMembers:
					userInfo.College = Homerton
				case HughesHallMembers:
					userInfo.College = HughesHall
				case JesusMembers:
					userInfo.College = Jesus
				case KingsMembers:
					userInfo.College = Kings
				case LucyCavendishMembers:
					userInfo.College = LucyCavendish
				case MagdaleneMembers:
					userInfo.College = Magdalene
				case MurrayEdwardsMembers:
					userInfo.College = MurrayEdwards
				case NewnhamMembers:
					userInfo.College = Newnham
				case PembrokeMembers:
					userInfo.College = Pembroke
				case PeterhouseMembers:
					userInfo.College = Peterhouse
				case QueensMembers:
					userInfo.College = Queens
				case RobinsonMembers:
					userInfo.College = Robinson
				case SelwynMembers:
					userInfo.College = Selwyn
				case SidneySussexMembers:
					userInfo.College = SidneySussex
				case StCatharinesMembers:
					userInfo.College = StCatharines
				case StEdmundsMembers:
					userInfo.College = StEdmunds
				case StJohnsMembers:
					userInfo.College = StJohns
				case TrinityMembers:
					userInfo.College = Trinity
				case TrinityHallMembers:
					userInfo.College = TrinityHall
				case WolfsonMembers:
					userInfo.College = Wolfson
				}
			}
		}
	}

	return userInfo, nil
}
