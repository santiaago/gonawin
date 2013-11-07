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

package auth

import (
	"net/http"
	
	usermdl "github.com/santiaago/purple-wing/models/user"
)

const kEmailRjourde = "remy.jourde@gmail.com"
const kEemailSarias = "santiago.ariassar@gmail.com"

func IsAuthorizedWithGoogle(ui *usermdl.GPlusUserInfo) bool {
	return ui != nil && (ui.Email == kEmailRjourde || ui.Email == kEemailSarias)
}

func IsAuthorizedWithTwitter(ui *usermdl.TwitterUserInfo) bool {
	return ui != nil && (ui.Screen_name == "rjourde" || ui.Screen_name == "santiago_arias")
}

func IsAuthorizedWithFacebook(ui *usermdl.FacebookUserInfo) bool {
	return ui != nil && (ui.Email == kEmailRjourde || ui.Email == kEemailSarias)
}

// LoggedIn is true is the AuthCookie exist and match your user.Auth property
func LoggedIn(r *http.Request) bool {
	if auth := GetAuthCookie(r); len(auth) > 0 {
		if u := CurrentUser(r); u != nil {
			return u.Auth == auth
		}
	}
	
	return false
}

// IsAdmin is true if you are logged in and belong to the below users.
func IsAdmin(r *http.Request) bool {
	if LoggedIn(r){
		if u := CurrentUser(r); u != nil{
			return (u.Email == "remy.jourde@gmail.com" || u.Email == "santiago.ariassar@gmail.com" || u.Username == "rjourde" || u.Username == "santiago_arias")
		}
	}
	return false
}

// IsUser is true if you are logged in, can either be an admin or not.
func IsUser(r *http.Request) bool {
	return LoggedIn(r)
}
