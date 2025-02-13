package coreapi

import (
	"context"
	"fmt"

	"github.com/ipfs/boxo/namesys/resolve"
	"github.com/ipfs/kubo/tracing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	coreiface "github.com/ipfs/boxo/coreiface"
	"github.com/ipfs/boxo/path"
	ipfspathresolver "github.com/ipfs/boxo/path/resolver"
	ipld "github.com/ipfs/go-ipld-format"
)

// ResolveNode resolves the path `p` using Unixfs resolver, gets and returns the
// resolved Node.
func (api *CoreAPI) ResolveNode(ctx context.Context, p path.Path) (ipld.Node, error) {
	ctx, span := tracing.Span(ctx, "CoreAPI", "ResolveNode", trace.WithAttributes(attribute.String("path", p.String())))
	defer span.End()

	rp, _, err := api.ResolvePath(ctx, p)
	if err != nil {
		return nil, err
	}

	node, err := api.dag.Get(ctx, rp.RootCid())
	if err != nil {
		return nil, err
	}
	return node, nil
}

// ResolvePath resolves the path `p` using Unixfs resolver, returns the
// resolved path.
func (api *CoreAPI) ResolvePath(ctx context.Context, p path.Path) (path.ImmutablePath, []string, error) {
	ctx, span := tracing.Span(ctx, "CoreAPI", "ResolvePath", trace.WithAttributes(attribute.String("path", p.String())))
	defer span.End()

	p, err := resolve.ResolveIPNS(ctx, api.namesys, p)
	if err == resolve.ErrNoNamesys {
		return nil, nil, coreiface.ErrOffline
	} else if err != nil {
		return nil, nil, err
	}

	var resolver ipfspathresolver.Resolver
	switch p.Namespace() {
	case path.IPLDNamespace:
		resolver = api.ipldPathResolver
	case path.IPFSNamespace:
		resolver = api.unixFSPathResolver
	default:
		return nil, nil, fmt.Errorf("unsupported path namespace: %s", p.Namespace())
	}

	imPath, err := path.NewImmutablePath(p)
	if err != nil {
		return nil, nil, err
	}

	node, remainder, err := resolver.ResolveToLastNode(ctx, imPath)
	if err != nil {
		return nil, nil, err
	}

	segments := []string{p.Namespace(), node.String()}
	segments = append(segments, remainder...)

	p, err = path.NewPathFromSegments(segments...)
	if err != nil {
		return nil, nil, err
	}

	imPath, err = path.NewImmutablePath(p)
	if err != nil {
		return nil, nil, err
	}

	return imPath, remainder, nil
}
