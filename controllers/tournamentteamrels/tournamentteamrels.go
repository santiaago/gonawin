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

package tournamentteamrels

import (
	"errors"
	"net/http"

	"appengine"

	"github.com/santiaago/purple-wing/helpers"
	"github.com/santiaago/purple-wing/helpers/handlers"
	"github.com/santiaago/purple-wing/helpers/log"
	templateshlp "github.com/santiaago/purple-wing/helpers/templates"

	tournamentmdl "github.com/santiaago/purple-wing/models/tournament"
	usermdl "github.com/santiaago/purple-wing/models/user"
)

// create handler for tournament teams realtionship
func CreateJson(w http.ResponseWriter, r *http.Request, u *usermdl.User) error {
	c := appengine.NewContext(r)

	if r.Method == "POST" {
		// get tournament and team id
		tournamentId, err1 := handlers.PermalinkID(r, c, 4)
		teamId, err2 := handlers.PermalinkID(r, c, 5)
		if err1 != nil || err2 != nil {
			log.Errorf(c, " tournamentteamrels.Create, string value could not be parsed: %v, %v", err1, err2)
			if err1 != nil {
				return helpers.NotFound{err1}
			} else if err2 != nil {
				return helpers.NotFound{err2}
			}
		}

		if err := tournamentmdl.TeamJoin(c, tournamentId, teamId); err != nil {
			log.Errorf(c, " tournamentteamrels.Create: %v", err)
			return helpers.InternalServerError{err}
		}
		// return the joined tournament
		if tournament, err := tournamentmdl.ById(c, tournamentId); err != nil {
			return helpers.NotFound{err}
		} else {
			var tJson tournamentmdl.TournamentJson
			fieldsToKeep := []string{"Id", "Name"}
			helpers.InitPointerStructure(tournament, &tJson, fieldsToKeep)

			return templateshlp.RenderJson(w, c, tJson)
		}
	} else {
		return helpers.BadRequest{errors.New("Not supported.")}
	}
}

// destroy handler for tournament teams realtionship
func DestroyJson(w http.ResponseWriter, r *http.Request, u *usermdl.User) error {
	c := appengine.NewContext(r)

	if r.Method == "POST" {

		// get tournament and team id
		tournamentId, err1 := handlers.PermalinkID(r, c, 4)
		teamId, err2 := handlers.PermalinkID(r, c, 5)
		if err1 != nil || err2 != nil {
			log.Errorf(c, " tournamentteamrels.Destroy, string value could not be parsed: %v, %v", err1, err2)
			if err1 != nil {
				return helpers.NotFound{err1}
			} else if err2 != nil {
				return helpers.NotFound{err2}
			}
		}
		// leave team
		if err := tournamentmdl.TeamLeave(c, tournamentId, teamId); err != nil {
			log.Errorf(c, " tournamentteamrels.Destroy: %v", err)
			return helpers.InternalServerError{err}
		}
		// return the left tournament
		if tournament, err := tournamentmdl.ById(c, tournamentId); err != nil {
			return helpers.NotFound{err}
		} else {
			var tJson tournamentmdl.TournamentJson
			fieldsToKeep := []string{"Id", "Name"}
			helpers.InitPointerStructure(tournament, &tJson, fieldsToKeep)

			return templateshlp.RenderJson(w, c, tJson)
		}
	} else {
		return helpers.BadRequest{errors.New("Not supported.")}
	}
}
