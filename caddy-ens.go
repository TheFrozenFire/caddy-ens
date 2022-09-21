package caddyens

import (
    "strings"
    "net/http"
    "encoding/hex"

    "github.com/ethereum/go-ethereum/ethclient"
    ens "github.com/wealdtech/go-ens/v3"
    multicodec "github.com/wealdtech/go-multicodec"
    
    "github.com/caddyserver/caddy/v2"
    "github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
    "github.com/caddyserver/caddy/v2/modules/caddyhttp"
    
    "go.uber.org/zap"
)

type EnsClient struct{
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
        ID: "ens",
        New: func() caddy.Module { return new(EnsClient) },
    }
}

func (c *EnsClient) Provision(ctx caddy.Context) error {
    c.logger = ctx.Logger()
}

//  caddy-ens {
//      
//  }
func (c *EnsClient) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
    for d.Next() {
        if d.NextArg() {
            return d.ArgErr()
        }
        
        for nesting := d.Nesting(); d.NextBlock(nesting); {
            switch d.Val() {
            case "eth_rpc_endpoint":
                if d.NextArg() {
                    c.client, err = ethclient.Dial(d.Val())
                    if err != nil {
                        panic(err)
                    }
                }
                if d.NextArg() {
                    return d.ArgErr()
                }
            default:
                return d.Errf("unrecognized subdirective '%s'", d.Val())
            }
        }
    }

    return nil
}

func (c *EnsClient) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
    resolver, err := ens.NewResolver(client, c.Domain)
    if err != nil  {
        panic(err)
    }
    
    logger.Debug("ENS domain resolver found",
        zap.String("domain", resolver.domain),
        zap.ByteString("resolver", ContractAddr)
    )
    
    headers := w.Header()
    
    for _, attributeName := range c.Attributes {
        attributeName = strings.ToLower(attributeName)
        
        switch attributeName {
        case "address":
            address, err := resolver.Address()
            if err != nil {
                panic(err)
            }
            
            logger.Debug("ENS domain address found",
                zap.String("domain", resolver.domain),
                zap.ByteString("address", address)
            )
            
            headers.Set("X-ENS-Address", hex.EncodeToString(address))
        case "contenthash":
            contentHash, err := resolver.Contenthash()
            if err != nil {
                panic(err)
            }
            
            logger.Debug("ENS domain content hash found",
                zap.String("domain", resolver.domain),
                zap.ByteString("contentHash", contentHash)
            )
            
            cH_address, cH_codec, err := multicodec.RemoveCodec(contentHash)
            
            headers.Set("X-ENS-Contenthash", hex.EncodeToString(contentHash))
            headers.Set("X-ENS-Contenthash-Codec", hex.EncodeToString(cH_codec))
            headers.Set("X-ENS-Contenthash-Address", hex.EncodeToString(cH_address))
        
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
