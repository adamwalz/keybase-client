// Auto-generated by avdl-compiler v1.3.11 (https://github.com/keybase/node-avdl-compiler)
//   Input file: avdl/keybase1/tlf_keys.avdl

package keybase1

import (
	"github.com/keybase/go-framed-msgpack-rpc/rpc"
	context "golang.org/x/net/context"
)

type TLFIdentifyBehavior int

const (
	TLFIdentifyBehavior_DEFAULT_KBFS    TLFIdentifyBehavior = 0
	TLFIdentifyBehavior_CHAT_CLI        TLFIdentifyBehavior = 1
	TLFIdentifyBehavior_CHAT_GUI        TLFIdentifyBehavior = 2
	TLFIdentifyBehavior_CHAT_GUI_STRICT TLFIdentifyBehavior = 3
	TLFIdentifyBehavior_KBFS_REKEY      TLFIdentifyBehavior = 4
	TLFIdentifyBehavior_KBFS_QR         TLFIdentifyBehavior = 5
)

var TLFIdentifyBehaviorMap = map[string]TLFIdentifyBehavior{
	"DEFAULT_KBFS":    0,
	"CHAT_CLI":        1,
	"CHAT_GUI":        2,
	"CHAT_GUI_STRICT": 3,
	"KBFS_REKEY":      4,
	"KBFS_QR":         5,
}

var TLFIdentifyBehaviorRevMap = map[TLFIdentifyBehavior]string{
	0: "DEFAULT_KBFS",
	1: "CHAT_CLI",
	2: "CHAT_GUI",
	3: "CHAT_GUI_STRICT",
	4: "KBFS_REKEY",
	5: "KBFS_QR",
}

func (e TLFIdentifyBehavior) String() string {
	if v, ok := TLFIdentifyBehaviorRevMap[e]; ok {
		return v
	}
	return ""
}

type CanonicalTlfName string
type CryptKey struct {
	KeyGeneration int     `codec:"KeyGeneration" json:"KeyGeneration"`
	Key           Bytes32 `codec:"Key" json:"Key"`
}

type TLFBreak struct {
	Breaks []TLFIdentifyFailure `codec:"breaks" json:"breaks"`
}

type TLFIdentifyFailure struct {
	User   User                 `codec:"user" json:"user"`
	Breaks *IdentifyTrackBreaks `codec:"breaks,omitempty" json:"breaks,omitempty"`
}

type CanonicalTLFNameAndIDWithBreaks struct {
	TlfID         TLFID            `codec:"tlfID" json:"tlfID"`
	CanonicalName CanonicalTlfName `codec:"CanonicalName" json:"CanonicalName"`
	Breaks        TLFBreak         `codec:"breaks" json:"breaks"`
}

type GetTLFCryptKeysRes struct {
	NameIDBreaks CanonicalTLFNameAndIDWithBreaks `codec:"nameIDBreaks" json:"nameIDBreaks"`
	CryptKeys    []CryptKey                      `codec:"CryptKeys" json:"CryptKeys"`
}

type TLFQuery struct {
	TlfName          string              `codec:"tlfName" json:"tlfName"`
	IdentifyBehavior TLFIdentifyBehavior `codec:"identifyBehavior" json:"identifyBehavior"`
}

type GetTLFCryptKeysArg struct {
	Query TLFQuery `codec:"query" json:"query"`
}

type GetPublicCanonicalTLFNameAndIDArg struct {
	Query TLFQuery `codec:"query" json:"query"`
}

type TlfKeysInterface interface {
	// getTLFCryptKeys returns TLF crypt keys from all generations and the TLF ID.
	// TLF ID should not be cached or stored persistently.
	GetTLFCryptKeys(context.Context, TLFQuery) (GetTLFCryptKeysRes, error)
	// getPublicCanonicalTLFNameAndID return the canonical name and TLFID for tlfName.
	// TLF ID should not be cached or stored persistently.
	GetPublicCanonicalTLFNameAndID(context.Context, TLFQuery) (CanonicalTLFNameAndIDWithBreaks, error)
}

func TlfKeysProtocol(i TlfKeysInterface) rpc.Protocol {
	return rpc.Protocol{
		Name: "keybase.1.tlfKeys",
		Methods: map[string]rpc.ServeHandlerDescription{
			"getTLFCryptKeys": {
				MakeArg: func() interface{} {
					ret := make([]GetTLFCryptKeysArg, 1)
					return &ret
				},
				Handler: func(ctx context.Context, args interface{}) (ret interface{}, err error) {
					typedArgs, ok := args.(*[]GetTLFCryptKeysArg)
					if !ok {
						err = rpc.NewTypeError((*[]GetTLFCryptKeysArg)(nil), args)
						return
					}
					ret, err = i.GetTLFCryptKeys(ctx, (*typedArgs)[0].Query)
					return
				},
				MethodType: rpc.MethodCall,
			},
			"getPublicCanonicalTLFNameAndID": {
				MakeArg: func() interface{} {
					ret := make([]GetPublicCanonicalTLFNameAndIDArg, 1)
					return &ret
				},
				Handler: func(ctx context.Context, args interface{}) (ret interface{}, err error) {
					typedArgs, ok := args.(*[]GetPublicCanonicalTLFNameAndIDArg)
					if !ok {
						err = rpc.NewTypeError((*[]GetPublicCanonicalTLFNameAndIDArg)(nil), args)
						return
					}
					ret, err = i.GetPublicCanonicalTLFNameAndID(ctx, (*typedArgs)[0].Query)
					return
				},
				MethodType: rpc.MethodCall,
			},
		},
	}
}

type TlfKeysClient struct {
	Cli rpc.GenericClient
}

// getTLFCryptKeys returns TLF crypt keys from all generations and the TLF ID.
// TLF ID should not be cached or stored persistently.
func (c TlfKeysClient) GetTLFCryptKeys(ctx context.Context, query TLFQuery) (res GetTLFCryptKeysRes, err error) {
	__arg := GetTLFCryptKeysArg{Query: query}
	err = c.Cli.Call(ctx, "keybase.1.tlfKeys.getTLFCryptKeys", []interface{}{__arg}, &res)
	return
}

// getPublicCanonicalTLFNameAndID return the canonical name and TLFID for tlfName.
// TLF ID should not be cached or stored persistently.
func (c TlfKeysClient) GetPublicCanonicalTLFNameAndID(ctx context.Context, query TLFQuery) (res CanonicalTLFNameAndIDWithBreaks, err error) {
	__arg := GetPublicCanonicalTLFNameAndIDArg{Query: query}
	err = c.Cli.Call(ctx, "keybase.1.tlfKeys.getPublicCanonicalTLFNameAndID", []interface{}{__arg}, &res)
	return
}
