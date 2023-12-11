package base

import (
	"sync"

	"github.com/luojinbo008/gost/common"
	"github.com/luojinbo008/gost/internal/cluster/router"
	"github.com/luojinbo008/gost/internal/cluster/router/chain"

	"go.uber.org/atomic"
)

// Directory Abstract implementation of Directory: Invoker list returned from this Directory's list method have been filtered by Routers
type Directory struct {
	url       *common.URL
	destroyed *atomic.Bool
	// this mutex for change the properties in BaseDirectory, like routerChain , destroyed etc
	mutex       sync.Mutex
	routerChain router.Chain
}

// NewDirectory Create BaseDirectory with URL
func NewDirectory(url *common.URL) Directory {
	return Directory{
		url:         url,
		destroyed:   atomic.NewBool(false),
		routerChain: &chain.RouterChain{},
	}
}

// RouterChain Return router chain in directory
func (dir *Directory) RouterChain() router.Chain {
	return dir.routerChain
}

// SetRouterChain Set router chain in directory
func (dir *Directory) SetRouterChain(routerChain router.Chain) {
	dir.mutex.Lock()
	defer dir.mutex.Unlock()
	dir.routerChain = routerChain
}

// GetURL Get URL
func (dir *Directory) GetURL() *common.URL {
	return dir.url
}

// GetDirectoryUrl Get URL instance
func (dir *Directory) GetDirectoryUrl() *common.URL {
	return dir.url
}

// func (dir *Directory) isProperRouter(url *common.URL) bool {
// 	app := url.GetParam(constant.ApplicationKey, "")
// 	dirApp := dir.GetURL().GetParam(constant.ApplicationKey, "")
// 	if len(dirApp) == 0 && dir.GetURL().SubURL != nil {
// 		dirApp = dir.GetURL().SubURL.GetParam(constant.ApplicationKey, "")
// 	}
// 	serviceKey := dir.GetURL().ServiceKey()
// 	if len(serviceKey) == 0 {
// 		serviceKey = dir.GetURL().SubURL.ServiceKey()
// 	}
// 	if len(app) > 0 && app == dirApp {
// 		return true
// 	}
// 	if url.ServiceKey() == serviceKey {
// 		return true
// 	}
// 	return false
// }

// Destroy Destroy
func (dir *Directory) Destroy(doDestroy func()) {
	if dir.destroyed.CAS(false, true) {
		dir.mutex.Lock()
		doDestroy()
		dir.mutex.Unlock()
	}
}

// IsAvailable Once directory init finish, it will change to true
func (dir *Directory) IsAvailable() bool {
	return !dir.destroyed.Load()
}
