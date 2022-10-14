# easycache

easycache is exactly what it sounds like, an easy to use cache that will automatically refetch from the origin while using stale data during the origin refetch.

Example:

```go
fun fetchUserFromDB(key any) any {
    var user User
    db.Find(&user, key)
    return user
}

func main() {

    // Setup cache
    userCache := easycache.New(ttl, fetchUserFromDB)

    // Pass in 'key' that `fetchUserFromDB` will use to reference.  `key` can be `any`thing :)
    uCache := userCache(1)

    // do a little type assertion, and done!
    user := uCache.(User)
}

```

### warning

1. easycache doesn't do any type checking, you must handle that on your end.  That allows you to use a single cache for any mixture of data you'd
2. easycache doesn't have a way of 'setting' or 'deleting' data on the fly, it's made to be easy
3. easycache doesn't do any memory management, it's simply pushing the data in a slice and handling lock management
4. easycache comes with no warranty, use at your own risk :)