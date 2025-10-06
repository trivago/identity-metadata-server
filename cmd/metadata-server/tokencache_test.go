package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAccessTokenLookupCollision(t *testing.T) {
	// Test the lookup collision
	assert := assert.New(t)

	fakeIdentity := MockSourceIdentity{
		Name:     "test",
		BoundGSA: "test@gcp.com",
	}

	accessTokenId1 := NewLookup(TokenTypeAccess, fakeIdentity)
	accessTokenId2 := NewLookupWithAudience(TokenTypeAccess, fakeIdentity, "audience")
	accessTokenId3 := NewLookupWithScopeAndAudience(TokenTypeAccess, fakeIdentity, []string{"scope"}, []string{"audience"})
	accessTokenId4 := NewLookupWithScopeAndAudience(TokenTypeAccess, fakeIdentity, []string{"scope", "scope2"}, []string{"audience"})
	accessTokenId5 := NewLookupWithScopeAndAudience(TokenTypeAccess, fakeIdentity, []string{"scope"}, []string{"audience", "audience2"})

	assert.NotEqual(accessTokenId1, accessTokenId2)
	assert.NotEqual(accessTokenId2, accessTokenId3)
	assert.NotEqual(accessTokenId3, accessTokenId4)
	assert.NotEqual(accessTokenId4, accessTokenId5)

	assert.NotEqual(accessTokenId1.ToTokenUID(), accessTokenId2.ToTokenUID())
	assert.NotEqual(accessTokenId2.ToTokenUID(), accessTokenId3.ToTokenUID())
	assert.NotEqual(accessTokenId3.ToTokenUID(), accessTokenId1.ToTokenUID())
	assert.NotEqual(accessTokenId4.ToTokenUID(), accessTokenId3.ToTokenUID())
	assert.NotEqual(accessTokenId5.ToTokenUID(), accessTokenId3.ToTokenUID())
}

func TestIdentityTokenLookupCollision(t *testing.T) {
	// Test the lookup collision
	assert := assert.New(t)

	fakeIdentity := MockSourceIdentity{
		Name:     "test",
		BoundGSA: "test@gcp.com",
	}

	identityTokenId1 := NewLookup(TokenTypeIdentity, fakeIdentity)
	identityTokenId2 := NewLookupWithAudience(TokenTypeIdentity, fakeIdentity, "audience")
	identityTokenId3 := NewLookupWithScopeAndAudience(TokenTypeIdentity, fakeIdentity, []string{"scope"}, []string{"audience"})
	identityTokenId4 := NewLookupWithScopeAndAudience(TokenTypeIdentity, fakeIdentity, []string{"scope", "scope2"}, []string{"audience"})
	identityTokenId5 := NewLookupWithScopeAndAudience(TokenTypeIdentity, fakeIdentity, []string{"scope"}, []string{"audience", "audience2"})

	assert.NotEqual(identityTokenId1.ToTokenUID(), identityTokenId2.ToTokenUID())
	assert.NotEqual(identityTokenId2.ToTokenUID(), identityTokenId3.ToTokenUID())
	assert.NotEqual(identityTokenId3.ToTokenUID(), identityTokenId1.ToTokenUID())
	assert.NotEqual(identityTokenId4.ToTokenUID(), identityTokenId3.ToTokenUID())
	assert.NotEqual(identityTokenId5.ToTokenUID(), identityTokenId3.ToTokenUID())
}

func TestCrossTokenLookupCollision(t *testing.T) {
	// Test the lookup collision
	assert := assert.New(t)

	fakeIdentity := MockSourceIdentity{
		Name:     "test",
		BoundGSA: "test@gcp.com",
	}

	accessTokenId1 := NewLookup(TokenTypeAccess, fakeIdentity)
	accessTokenId2 := NewLookupWithAudience(TokenTypeAccess, fakeIdentity, "audience")
	accessTokenId3 := NewLookupWithScopeAndAudience(TokenTypeAccess, fakeIdentity, []string{"scope"}, []string{"audience"})
	identityTokenId1 := NewLookup(TokenTypeIdentity, fakeIdentity)
	identityTokenId2 := NewLookupWithAudience(TokenTypeIdentity, fakeIdentity, "audience")
	identityTokenId3 := NewLookupWithScopeAndAudience(TokenTypeIdentity, fakeIdentity, []string{"scope"}, []string{"audience"})

	assert.NotEqual(accessTokenId1.ToTokenUID(), identityTokenId1.ToTokenUID())
	assert.NotEqual(accessTokenId2.ToTokenUID(), identityTokenId2.ToTokenUID())
	assert.NotEqual(accessTokenId3.ToTokenUID(), identityTokenId3.ToTokenUID())
}

func TestTokenCacheSetGet(t *testing.T) {
	// Test the token cache
	assert := assert.New(t)

	// Create a new cache
	cache := NewTokenCache(0, 0)

	fakeIdentity := MockSourceIdentity{
		Name:     "test",
		BoundGSA: "test@gcp.com",
	}

	// Test access token

	accessTokenId1 := NewLookup(TokenTypeAccess, fakeIdentity)
	accessTokenId2 := NewLookupWithAudience(TokenTypeAccess, fakeIdentity, "audience")
	accessTokenId3 := NewLookupWithScopeAndAudience(TokenTypeAccess, fakeIdentity, []string{"scope"}, []string{"audience"})

	cache.Store(accessTokenId1, "test1", time.Now().Add(time.Minute))
	cache.Store(accessTokenId2, "test2", time.Now().Add(time.Minute))
	cache.Store(accessTokenId3, "test3", time.Now().Add(time.Minute))

	t1 := cache.Get(accessTokenId1)
	t2 := cache.Get(accessTokenId2)
	t3 := cache.Get(accessTokenId3)

	assert.NotNil(t1)
	assert.NotNil(t2)
	assert.NotNil(t3)

	assert.Equal("test1", t1.token)
	assert.Equal("test2", t2.token)
	assert.Equal("test3", t3.token)

	// Test identity token

	identityTokenId1 := NewLookup(TokenTypeIdentity, fakeIdentity)
	identityTokenId2 := NewLookupWithAudience(TokenTypeIdentity, fakeIdentity, "audience")
	identityTokenId3 := NewLookupWithScopeAndAudience(TokenTypeIdentity, fakeIdentity, []string{"scope"}, []string{"audience"})

	cache.Store(identityTokenId1, "idTest1", time.Now().Add(time.Minute))
	cache.Store(identityTokenId2, "idTest2", time.Now().Add(time.Minute))
	cache.Store(identityTokenId3, "idTest3", time.Now().Add(time.Minute))

	t1 = cache.Get(identityTokenId1)
	t2 = cache.Get(identityTokenId2)
	t3 = cache.Get(identityTokenId3)

	assert.NotNil(t1)
	assert.NotNil(t2)
	assert.NotNil(t3)

	assert.Equal("idTest1", t1.token)
	assert.Equal("idTest2", t2.token)
	assert.Equal("idTest3", t3.token)
}

func TestTokenCacheInvalidation(t *testing.T) {
	// Test the token cache
	assert := assert.New(t)

	// Create a new cache
	cache := NewTokenCache(0, 0)

	fakeIdentity := MockSourceIdentity{
		Name:     "test",
		BoundGSA: "test@gcp.com",
	}

	// Test access token

	lifeTime := 200 * time.Millisecond
	accessTokenId1 := NewLookup(TokenTypeAccess, fakeIdentity)
	identityTokenId1 := NewLookup(TokenTypeIdentity, fakeIdentity)

	cache.Store(accessTokenId1, "test1", time.Now().Add(lifeTime))
	cache.Store(identityTokenId1, "test2", time.Now().Add(lifeTime))

	time.Sleep(lifeTime)

	assert.Nil(cache.Get(accessTokenId1))
	assert.Nil(cache.Get(identityTokenId1))

	assert.NotPanics(func() { cache.Get(accessTokenId1) })
	assert.NotPanics(func() { cache.Get(identityTokenId1) })
}

func TestTokenMinLifetime(t *testing.T) {
	// Test the token cache
	assert := assert.New(t)

	// Create a new cache with a minimum token lifetime of 5 seconds
	// This means that a token will be removed from the cache if it has less than 5 seconds left
	cache := NewTokenCache(0, 5*time.Second)

	fakeIdentity := MockSourceIdentity{
		Name:     "test",
		BoundGSA: "test@gcp.com",
	}

	// Test access token with a lifetime below the minimum

	lifeTime := time.Second
	accessTokenId1 := NewLookup(TokenTypeAccess, fakeIdentity)
	identityTokenId1 := NewLookup(TokenTypeIdentity, fakeIdentity)

	cache.Store(accessTokenId1, "test1", time.Now().Add(lifeTime))
	cache.Store(identityTokenId1, "test2", time.Now().Add(lifeTime))

	// The token is immediately removed from the cache as it has less than 5 seconds left

	assert.Nil(cache.Get(accessTokenId1))
	assert.Nil(cache.Get(identityTokenId1))

	assert.NotPanics(func() { cache.Get(accessTokenId1) })
	assert.NotPanics(func() { cache.Get(identityTokenId1) })
}

func TestTokenCacheGC(t *testing.T) {
	// Test the token cache
	assert := assert.New(t)
	lifeTime := 200 * time.Millisecond

	// Create a new cache
	cache := NewTokenCache(lifeTime/2, 0)
	defer cache.StopGC()

	fakeIdentity := MockSourceIdentity{
		Name:     "test",
		BoundGSA: "test@gcp.com",
	}

	// Test access token

	accessTokenId1 := NewLookup(TokenTypeAccess, fakeIdentity)
	identityTokenId1 := NewLookup(TokenTypeIdentity, fakeIdentity)

	cache.Store(accessTokenId1, "test1", time.Now().Add(lifeTime))
	cache.Store(identityTokenId1, "test2", time.Now().Add(lifeTime))

	time.Sleep(lifeTime * 2)

	assert.Nil(cache.Get(accessTokenId1))
	assert.Nil(cache.Get(identityTokenId1))
}
