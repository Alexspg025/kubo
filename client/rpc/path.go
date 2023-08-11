package rpc

import (
	"context"

	"github.com/ipfs/boxo/path"
	cid "github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
)

func (api *HttpApi) ResolvePath(ctx context.Context, p path.Path) (path.ImmutablePath, error) {
	var out struct {
		Cid     cid.Cid
		RemPath string
	}

	// TODO: this is hacky, fixing https://github.com/ipfs/go-ipfs/issues/5703 would help

	var err error
	if p.Namespace() == path.IPNSNamespace {
		if p, err = api.Name().Resolve(ctx, p.String()); err != nil {
			return nil, err
		}
	}

	if err := api.Request("dag/resolve", p.String()).Exec(ctx, &out); err != nil {
		return nil, err
	}

	p, err = path.NewPathFromSegments(p.Namespace().String(), out.Cid.String(), out.RemPath)
	if err != nil {
		return nil, err
	}

	return path.NewImmutablePath(p)
}

func (api *HttpApi) ResolveNode(ctx context.Context, p path.Path) (ipld.Node, error) {
	rp, err := api.ResolvePath(ctx, p)
	if err != nil {
		return nil, err
	}

	return api.Dag().Get(ctx, rp.Cid())
}
