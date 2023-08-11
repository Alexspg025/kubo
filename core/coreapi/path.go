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

	rp, err := api.ResolvePath(ctx, p)
	if err != nil {
		return nil, err
	}

	node, err := api.dag.Get(ctx, rp.Cid())
	if err != nil {
		return nil, err
	}
	return node, nil
}

// ResolvePath resolves the path `p` using Unixfs resolver, returns the
// resolved path.
func (api *CoreAPI) ResolvePath(ctx context.Context, p path.Path) (path.ImmutablePath, error) {
	ctx, span := tracing.Span(ctx, "CoreAPI", "ResolvePath", trace.WithAttributes(attribute.String("path", p.String())))
	defer span.End()

	p, err := resolve.ResolveIPNS(ctx, api.namesys, p)
	if err == resolve.ErrNoNamesys {
		return nil, coreiface.ErrOffline
	} else if err != nil {
		return nil, err
	}

	if p.Namespace() != path.IPFSNamespace && p.Namespace() != path.IPLDNamespace {
		return nil, fmt.Errorf("unsupported path namespace: %s", p.Namespace().String())
	}

	var resolver ipfspathresolver.Resolver
	if p.Namespace() == path.IPLDNamespace {
		resolver = api.ipldPathResolver
	} else {
		resolver = api.unixFSPathResolver
	}

	node, rest, err := resolver.ResolveToLastNode(ctx, p)
	if err != nil {
		return nil, err
	}

	segments := []string{p.Namespace().String(), node.String()}
	segments = append(segments, rest...)

	p, err = path.NewPathFromSegments(segments...)
	if err != nil {
		return nil, err
	}

	return path.NewImmutablePath(p)
}
