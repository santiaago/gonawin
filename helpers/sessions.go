/*
 * Copyright (c) 2013 Santiago Arias | Remy Jourde | Carlos Bernal
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

package helpers

import (
    "net/http"
    "io"
    "fmt"
    "crypto/rand"
    
    "appengine"
    "appengine/memcache"
    
	"github.com/santiaago/purple-wing/models"
)

func IsAuthorized(ui *models.GPlusUserInfo) bool {
	return ui != nil && (ui.Email == "remy.jourde@gmail.com" || ui.Email == "santiago.ariassar@gmail.com")
}

func StoreAuthKey(r *http.Request, uid int64, auth string) {
    c := appengine.NewContext(r)
    
    item := &memcache.Item{
        Key:   "auth:"+auth,
        Value: []byte(fmt.Sprintf("%d", uid)),
    }
    // Set the item, unconditionally
    if err := memcache.Set(c, item); err != nil {
        c.Errorf("pw: error setting item: %v", err)
    }
}

func FetchAuthKey(r *http.Request, auth string) string {
    c := appengine.NewContext(r)

    // Get the item from the memcache
    if item, err := memcache.Get(c, "auth:"+auth); err == nil {
        return string(item.Value)
    } 
    
    return ""
}

func SetAuthCookie(w http.ResponseWriter, auth string) {
    cookie := &http.Cookie{ 
        Name: "auth", 
        Value: auth, 
        Path: "/", 
    }
    http.SetCookie(w, cookie)
}

func GetAuthCookie(r *http.Request) string {
    if cookie, err := r.Cookie("auth"); err == nil {
        return cookie.Value
    }

    return ""
}

func GenerateAuthKey() string {
    b := make([]byte, 32)
    if _, err := io.ReadFull(rand.Reader, b); err != nil {
        return ""
    }
    return string(b)
}
