package caddyens

import (
    "strings"
    "net/http"
    "encoding/hex"
    "fmt"

    "github.com/ethereum/go-ethereum/ethclient"
    ens "github.com/wealdtech/go-ens/v3"
    multicodec "github.com/wealdtech/go-multicodec"
    
    cid "github.com/ipfs/go-cid"
    
    "github.com/caddyserver/caddy/v2"
    "github.com/caddyserver/caddy/v2/modules/caddyhttp"
    
    "go.uber.org/zap"
)

type EnsClient struct{
    EthRpcEndpoint string `json:"eth_rpc_endpoint,omitempty"`

    Domain string `json:"domain,omitempty"`
    
    Attributes []string `json:"attributes,omitempty"`

    client *ethclient.Client
    logger *zap.Logger
}

func init() {
    caddy.RegisterModule(EnsClient{})
}

// CaddyModule returns the Caddy module information.
func (EnsClient) CaddyModule() caddy.ModuleInfo {
    return caddy.ModuleInfo{
        ID: "http.handlers.ens",
        New: func() caddy.Module { return new(EnsClient) },
    }
}

func (c *EnsClient) Provision(ctx caddy.Context) error {
    c.logger = ctx.Logger(c)
    
    client, err := ethclient.Dial(c.EthRpcEndpoint)
    if err != nil {
        panic(err)
    }
    c.client = client
    
    return nil
}

func decodeContentHash(contentHash []byte) (string, string, error) {
    address_data, codec, err := multicodec.RemoveCodec(contentHash)
    
    if err != nil {
        return "", "", err
    }
    
    codec_name, err := multicodec.Name(codec)
    
    if err != nil {
        return "", "", err
    }
    
    switch(codec_name) {
    case "ipfs-ns":
    case "ipns-ns":
        cid, err := cid.Cast(address_data)
        
        if err != nil {
            return codec_name, "", err
        }
        
        address := cid.String()
        return codec_name, address, nil
    }
    
    address := hex.EncodeToString(address_data)
    return codec_name, address, nil
}

func (c EnsClient) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
    resolver, err := ens.NewResolver(c.client, c.Domain)
    if err != nil  {
        panic(err)
    }
    
    c.logger.Debug("ENS domain resolver found", zap.String("domain", c.Domain), zap.String("resolver", resolver.ContractAddr.String()) )
    
    headers := w.Header()
    
    for _, attributeName := range c.Attributes {
        attributeName = strings.ToLower(attributeName)
        
        switch attributeName {
        case "address":
            address, err := resolver.Address()
            if err != nil {
                panic(err)
            }
            
            c.logger.Debug("ENS domain address found", zap.String("domain", c.Domain), zap.String("address", address.String()) )
            
            headers.Set("X-ENS-Address", address.String())
        case "contenthash":
            contentHash, err := resolver.Contenthash()
            if err != nil {
                panic(err)
            }
            
            c.logger.Debug("ENS domain content hash found", zap.String("domain", c.Domain), zap.String("contentHash", hex.EncodeToString(contentHash)) )
            
            codec, address, err := decodeContentHash(contentHash)
            if err != nil {
                panic(err)
            }
            
            headers.Set("X-ENS-Contenthash", hex.EncodeToString(contentHash))
            headers.Set("X-ENS-Contenthash-Codec", codec)
            headers.Set("X-ENS-Contenthash-Address", address)
        case "public_key":
            pubKey_x, pubKey_y, err := resolver.PubKey()
            if err != nil {
                panic(err)
            }
            
            c.logger.Debug("ENS domain public key found", zap.String("pubKey_x", hex.EncodeToString(pubKey_x[:])), zap.String("pubKey_y", hex.EncodeToString(pubKey_y[:])))
            
            headers.Set("X-ENS-Public-Key-X", hex.EncodeToString(pubKey_x[:]))
            headers.Set("X-ENS-Public-Key-Y", hex.EncodeToString(pubKey_y[:]))
        case "resolver_address":
            headers.Set("X-ENS-Resolver-Address", resolver.ContractAddr.String())
        default:
            return fmt.Errorf("unrecognized ENS attribute '%s'", attributeName)
        }
    }

    
    
    return next.ServeHTTP(w, r)
}

var (
    _ caddy.Provisioner = (*EnsClient)(nil)
    _ caddyhttp.MiddlewareHandler = (*EnsClient)(nil)
)
