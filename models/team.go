/*
 * Copyright (c) 2014 Santiago Arias | Remy Jourde
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

package models

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"appengine"
	"appengine/datastore"

	"github.com/taironas/gonawin/helpers"
	"github.com/taironas/gonawin/helpers/log"
)

// AccOfTournaments holds a tournament id and an accuracy id.
//
type AccOfTournaments struct {
	AccuracyId   int64 // id of accuracy entity
	TournamentId int64 // id of tournament
}

// Team holds tournament entity data.
//
type Team struct {
	Id               int64
	KeyName          string
	Name             string
	Description      string
	AdminIds         []int64 // ids of User that are admins of the team
	Private          bool
	Created          time.Time
	UserIds          []int64            // ids of Users <=> members of the team.
	TournamentIds    []int64            // ids of Tournaments <=> Tournaments the team subscribed.
	Accuracy         float64            // Overall Team accuracy.
	AccOfTournaments []AccOfTournaments // ids of Accuracies for each tournament the team is participating on .
	PriceIds         []int64            // ids of Prices <=> prices defined for each tournament the team participates.
	MembersCount     int64              // number of members in team
}

// TeamJSON is the JSON version of the Team struct.
//
type TeamJSON struct {
	Id               *int64              `json:",omitempty"`
	KeyName          *string             `json:",omitempty"`
	Name             *string             `json:",omitempty"`
	Description      *string             `json:",omitempty"`
	AdminIds         *[]int64            `json:",omitempty"`
	Private          *bool               `json:",omitempty"`
	Created          *time.Time          `json:",omitempty"`
	UserIds          *[]int64            `json:",omitempty"`
	TournamentIds    *[]int64            `json:",omitempty"`
	Accuracy         *float64            `json:",omitempty"`
	AccOfTournaments *[]AccOfTournaments `json:",omitempty"`
	PriceIds         *[]int64            `json:",omitempty"`
	MembersCount     *int64              `json:",omitempty"`
}

// CreateTeam creates a team given a name, description, an admin id and a private mode.
//
func CreateTeam(c appengine.Context, name string, description string, adminID int64, private bool) (*Team, error) {
	// create new team
	teamID, _, err := datastore.AllocateIDs(c, "Team", nil, 1)
	if err != nil {
		return nil, err
	}

	key := datastore.NewKey(c, "Team", "", teamID, nil)
	admins := make([]int64, 1)
	admins[0] = adminID
	var emptyArray []int64
	var emtpyArrayOfAccOfTournament []AccOfTournaments
	team := &Team{teamID, helpers.TrimLower(name), name, description, admins, private, time.Now(), emptyArray, emptyArray, float64(0), emtpyArrayOfAccOfTournament, emptyArray, 0}

	_, err = datastore.Put(c, key, team)
	if err != nil {
		return nil, err
	}
	// udpate inverted index
	AddToTeamInvertedIndex(c, helpers.TrimLower(name), teamID)

	return team, err
}

// Destroy a team given a team id.
//
func (t *Team) Destroy(c appengine.Context) error {

	if _, err := TeamByID(c, t.Id); err != nil {
		return fmt.Errorf("Cannot find team with Id=%d", t.Id)
	}

	key := datastore.NewKey(c, "Team", "", t.Id, nil)
	if errd := datastore.Delete(c, key); errd != nil {
		return errd
	}

	// remove key name.
	return UpdateTeamInvertedIndex(c, t.KeyName, "", t.Id)
}

// FindTeams searches for all Team entities with respect of a filter and a value.
//
func FindTeams(c appengine.Context, filter string, value interface{}) []*Team {

	q := datastore.NewQuery("Team").Filter(filter+" =", value)

	var teams []*Team

	if _, err := q.GetAll(c, &teams); err != nil {
		log.Errorf(c, " Team.Find, error occurred during GetAll: %v", err)
		return nil
	}

	return teams
}

// TeamByID gets a team given an id.
//
func TeamByID(c appengine.Context, Id int64) (*Team, error) {

	var t Team
	key := datastore.NewKey(c, "Team", "", Id, nil)

	if err := datastore.Get(c, key, &t); err != nil {
		log.Errorf(c, " team not found : %v", err)
		return nil, err
	}
	return &t, nil
}

// TeamKeyByID gets a team key given a team id.
//
func TeamKeyByID(c appengine.Context, Id int64) *datastore.Key {
	return datastore.NewKey(c, "Team", "", Id, nil)
}

// Update updates a team given an id and a team pointer.
//
func (t *Team) Update(c appengine.Context) error {
	// update key name
	t.KeyName = helpers.TrimLower(t.Name)
	k := TeamKeyByID(c, t.Id)
	oldTeam := new(Team)
	var err error
	if err = datastore.Get(c, k, oldTeam); err == nil {
		if _, err = datastore.Put(c, k, t); err != nil {
			return err
		}
		// use lower trim names as team inverted index store them like this.
		UpdateTeamInvertedIndex(c, oldTeam.KeyName, t.KeyName, t.Id)
	}
	return err
}

// FindAllTeams gets all teams in datastore.
//
func FindAllTeams(c appengine.Context) []*Team {
	q := datastore.NewQuery("Team")

	var teams []*Team

	if _, err := q.GetAll(c, &teams); err != nil {
		log.Errorf(c, " Team.FindAll, error occurred during GetAll call: %v", err)
	}

	return teams
}

// GetNotJoinedTeams gets all teams that a user has not joined
// with respect to the count and page.
//
func GetNotJoinedTeams(c appengine.Context, u *User, count, page int64) []*Team {
	desc := "Get not joined teams"
	teams := FindAllTeams(c)

	var notJoined []*Team
	for _, team := range teams {
		if !team.Joined(c, u) {
			notJoined = append(notJoined, team)
		}
	}
	// loop backward on all of these ids to fetch the teams
	log.Infof(c, "%s calculateStartAndEnd(%v, %v, %v)", desc, int64(len(notJoined)), count, page)
	start, end := calculateStartAndEnd(int64(len(notJoined)), count, page)

	log.Infof(c, "%s start = %d, end = %d", desc, start, end)

	var paged []*Team
	for i := start; i >= end; i-- {
		paged = append(paged, notJoined[i])
	}

	return paged
}

// TeamsByIDs returns an array of teams from a given team IDs array.
// An error could be returned.
//
func TeamsByIDs(c appengine.Context, IDs []int64) ([]*Team, error) {

	teams := make([]Team, len(IDs))
	keys := TeamsKeysByIDs(c, IDs)

	var wrongIndexes []int

	if err := datastore.GetMulti(c, keys, teams); err != nil {
		if me, ok := err.(appengine.MultiError); ok {
			for i, merr := range me {
				if merr == datastore.ErrNoSuchEntity {
					log.Errorf(c, "TeamsByIDs, missing key: %v %v", err, keys[i].IntID())

					wrongIndexes = append(wrongIndexes, i)
				}
			}
		} else {
			return nil, err
		}
	}

	var existingTeams []*Team
	for i := range teams {
		if !contains(wrongIndexes, i) {
			log.Infof(c, "TeamsByIDs %v", teams[i])
			existingTeams = append(existingTeams, &teams[i])
		}
	}
	return existingTeams, nil
}

// TeamsKeysByIDs returns an array of datastore keys from a given team IDs array.
//
func TeamsKeysByIDs(c appengine.Context, IDs []int64) []*datastore.Key {
	keys := make([]*datastore.Key, len(IDs))
	for i, id := range IDs {
		keys[i] = TeamKeyByID(c, id)
	}
	return keys
}

// Joined checks if a user has joined a team or not.
//
func (t *Team) Joined(c appengine.Context, u *User) bool {

	hasTeam, _ := u.ContainsTeamID(t.Id)
	return hasTeam
}

// Join let a user join a team.
// TeamId is added to user entity.
// UserId is added to team entity.
// UserId is added to all current tournaments joined by the team entity.
//
func (t *Team) Join(c appengine.Context, u *User) error {
	if err := u.AddTeamID(c, t.Id); err != nil {
		return fmt.Errorf(" Team.Join, error joining team for user:%d Error: %v", u.Id, err)
	}

	if err := t.AddUserID(c, u.Id); err != nil {
		return fmt.Errorf(" Team.Join, error joining team for user:%d Error: %v", u.Id, err)
	}

	if err := t.AddUserToTournaments(c, u.Id); err != nil {
		return fmt.Errorf("Team.Join, error adding user:%d to teams tournaments Error: %v", u.Id, err)
	}
	return nil
}

// Leave makes a user leave a team.
// Todo: Should we check that the user is indeed a member of the team?
//
func (t *Team) Leave(c appengine.Context, u *User) error {
	if err := u.RemoveTeamID(c, t.Id); err != nil {
		return fmt.Errorf(" Team.Leave, error leaving team for user:%v Error: %v", u.Id, err)
	}
	if err := t.RemoveUserID(c, u.Id); err != nil {
		return fmt.Errorf(" Team.Leave, error leaving team for user:%v Error: %v", u.Id, err)
	}
	// sar: 7 mar 2014
	// when a user leaves a team should we unsubscribe him from the tournaments of that team?
	// for now I would say no.

	// if err := t.RemoveUserFromTournaments(c, u); err != nil{
	// 	return fmt.Errorf("Team.Leave, error leaving teams tournaments for user:%v Error: %v", u.Id, err)
	// }
	return nil
}

// IsTeamAdmin checks if user is admin of the team with id 'teamId'.
//
func IsTeamAdmin(c appengine.Context, teamID int64, userID int64) bool {

	var team *Team
	if team, err := TeamByID(c, teamID); team == nil || err != nil {
		log.Errorf(c, " Team.IsTeamAdmin, error occurred during ById call: %v", err)
		return false
	}

	for _, aid := range team.AdminIds {
		if aid == userID {
			return true
		}
	}
	return false
}

// GetWordFrequencyForTeam will get the frequency of that word in the team terms
// given a id, and a word.
//
func GetWordFrequencyForTeam(c appengine.Context, Id int64, word string) int64 {

	if teams := FindTeams(c, "Id", Id); teams != nil {
		return helpers.CountTerm(strings.Split(teams[0].KeyName, " "), word)
	}
	return 0
}

// Players returns an array of users/ players that participates the given team.
//
func (t *Team) Players(c appengine.Context) ([]*User, error) {

	var users []*User
	var err error
	if users, err = UsersByIds(c, t.UserIds); err != nil {
		return nil, err
	}

	return users, err
}

// ContainsTournamentID checks if a given tournament id exists in a team entity.
//
func (t *Team) ContainsTournamentID(id int64) (bool, int) {

	for i, tID := range t.TournamentIds {
		if tID == id {
			return true, i
		}
	}
	return false, -1
}

// ContainsPriceID checks if a given price id exists in a team entity.
//
func (t *Team) ContainsPriceID(id int64) (bool, int) {
	for i, pID := range t.PriceIds {
		if pID == id {
			return true, i
		}
	}
	return false, -1

}

// AddTournamentID adds a tournament Id in the TournamentId array.
//
func (t *Team) AddTournamentID(c appengine.Context, tID int64) error {
	if hasTournament, _ := t.ContainsTournamentID(tID); hasTournament {
		return fmt.Errorf("AddTournamentID, allready a member.")
	}

	t.TournamentIds = append(t.TournamentIds, tID)
	if err := t.Update(c); err != nil {
		return err
	}

	for _, uID := range t.UserIds {
		user, err := UserByID(c, uID)
		if err != nil {
			log.Errorf(c, "Team.AddTournamentID, user not found")
		} else {
			log.Infof(c, "team Add tournament id add tournament id%v", user.Id)
			user.AddTournamentID(c, tID)
		}
	}
	return nil
}

// RemoveTournamentID removes a tournament Id in the TournamentId array.
//
func (t *Team) RemoveTournamentID(c appengine.Context, tID int64) error {

	hasTournament := false
	i := 0

	if hasTournament, i = t.ContainsTournamentID(tID); !hasTournament {
		return fmt.Errorf("RemoveTournamentID, not a member.")
	}

	// as the order of index in tournamentsId is not important,
	// replace elem at index i with last element and resize slice.
	t.TournamentIds[i] = t.TournamentIds[len(t.TournamentIds)-1]
	t.TournamentIds = t.TournamentIds[0 : len(t.TournamentIds)-1]

	if err := t.Update(c); err != nil {
		return err
	}
	return nil
}

// AddPriceID adds a tournament Id in the TournamentId array.
//
func (t *Team) AddPriceID(c appengine.Context, pID int64) error {

	if hasPrice, _ := t.ContainsPriceID(pID); hasPrice {
		return fmt.Errorf("AddPriceId, allready a member.")
	}

	t.PriceIds = append(t.PriceIds, pID)
	if err := t.Update(c); err != nil {
		return err
	}
	return nil
}

// removePriceID removes a price Id in the PriceID array.
//
func (t *Team) removePriceID(c appengine.Context, pID int64) error {

	hasPrice := false
	i := 0

	if hasPrice, i = t.ContainsPriceID(pID); !hasPrice {
		return fmt.Errorf("removePriceId, not a member")
	}

	t.PriceIds[i] = t.PriceIds[len(t.PriceIds)-1]
	t.PriceIds = t.PriceIds[0 : len(t.PriceIds)-1]

	if err := t.Update(c); err != nil {
		return err
	}
	return nil
}

// RemovePriceByTournamentID removes price enity and price id from team enity with respect to tournament id.
//
func (t *Team) RemovePriceByTournamentID(c appengine.Context, tID int64) error {
	for _, pID := range t.PriceIds {
		if p, err := PriceByID(c, pID); err == nil {
			if p.TeamId == t.Id && p.TournamentId == tID {
				if err1 := t.removePriceID(c, p.Id); err1 != nil {
					return err1
				}
				if err1 := p.Destroy(c); err1 != nil {
					return err1
				}
				return nil
			}
		}
	}
	return fmt.Errorf("RemovePriceByTournamentId price id not found. Team: %d tournament:%d", t.Id, tID)
}

// Tournaments return an array of tournament the team is involved in.
//
func (t *Team) Tournaments(c appengine.Context) []*Tournament {

	var tournaments []*Tournament
	var err error

	if tournaments, err = TournamentsByIds(c, t.TournamentIds); err != nil {
		log.Errorf(c, "Something failed when calling TournamentsByIDs from team.Tournaments: %v", err)
	}

	return tournaments
}

// Accuracies returns an array of accuracy the user is involved in.
//
func (t *Team) Accuracies(c appengine.Context) []*Accuracy {

	var accs []*Accuracy

	for _, acc := range t.AccOfTournaments {
		a, err := AccuracyByID(c, acc.AccuracyId)
		if err != nil {
			log.Errorf(c, " Accuracies, cannot find accuracy with Id=%", acc.AccuracyId)
		} else {
			accs = append(accs, a)
		}
	}
	return accs
}

// RemoveUserID removes a user Id in the UserId array.
//
func (t *Team) RemoveUserID(c appengine.Context, uID int64) error {

	hasUser := false
	i := 0

	if hasUser, i = t.ContainsUserID(uID); !hasUser {
		return fmt.Errorf("RemoveUserID, not a member")
	}

	// as the order of index in usersId is not important,
	// replace elem at index i with last element and resize slice.
	t.UserIds[i] = t.UserIds[len(t.UserIds)-1]
	t.UserIds = t.UserIds[0 : len(t.UserIds)-1]
	t.MembersCount = int64(len(t.UserIds))

	if err := t.Update(c); err != nil {
		return err
	}
	return nil
}

// AddUserID adds a team Id in the UserId array.
//
func (t *Team) AddUserID(c appengine.Context, uID int64) error {

	if hasUser, _ := t.ContainsUserID(uID); hasUser {
		return fmt.Errorf("AddUserID, allready a member")
	}

	t.UserIds = append(t.UserIds, uID)
	t.MembersCount = int64(len(t.UserIds))
	if err := t.Update(c); err != nil {
		return err
	}
	return nil
}

// AddUserToTournaments add user to teams current tournaments.
//
func (t *Team) AddUserToTournaments(c appengine.Context, uID int64) error {
	if len(t.TournamentIds) == 0 {
		return nil
	}

	var u *User
	var err error

	if u, err = UserByID(c, uID); err != nil {
		log.Errorf(c, "User not found %v", uID)
		return err
	}

	for _, tID := range t.TournamentIds {
		if tournament, err := TournamentByID(c, tID); err != nil {
			log.Errorf(c, "Cannot find tournament with Id=%d", tID)
		} else if time.Now().Before(tournament.End) {

			if err := tournament.AddUserID(c, uID); err != nil {
				log.Errorf(c, "Team.AddUserToTournaments: unable to add user:%d to tournament:%d", uID, tID)
			}

			if err = u.AddTournamentID(c, tournament.Id); err != nil {
				log.Errorf(c, "Team.AddUserToTournaments: unable to add tournament id:%d to user:%d, %v", tID, uID, err)
			}
		}
	}
	return nil
}

// AddAdmin adds user to admins of current team.
// In order to be an admin of a team you should first be a member of the team.
//
func (t *Team) AddAdmin(c appengine.Context, id int64) error {

	if isMember, _ := t.ContainsUserID(id); isMember {
		if isAdmin, _ := t.ContainsAdminID(id); isAdmin {
			return fmt.Errorf("User with %d is already an admin of team", id)
		}
		t.AdminIds = append(t.AdminIds, id)
		if err := t.Update(c); err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("User with %d is not a member of the team", id)
}

// RemoveAdmin remove user of admins array in current team.
// In order to remove an admin from a team, there should be at least an admin in the array.
//
func (t *Team) RemoveAdmin(c appengine.Context, id int64) error {

	if isMember, _ := t.ContainsUserID(id); isMember {
		if isAdmin, i := t.ContainsAdminID(id); isAdmin {
			if len(t.AdminIds) > 1 {
				// as the order of index in adminIds is not important,
				// replace elem at index i with last element and resize slice.
				t.AdminIds[i] = t.AdminIds[len(t.AdminIds)-1]
				t.AdminIds = t.AdminIds[0 : len(t.AdminIds)-1]
				if err := t.Update(c); err != nil {
					return err
				}
				return nil
			}
			return fmt.Errorf("Cannot remove admin %d as there are no admins left in team", id)
		}
		return fmt.Errorf("User with %d is not admin of the team", id)
	}
	return fmt.Errorf("User with %d is not a member of the team", id)
}

// ContainsAdminID checks if user is admin of team.
//
func (t *Team) ContainsAdminID(id int64) (bool, int) {

	for i, aID := range t.AdminIds {
		if aID == id {
			return true, i
		}
	}
	return false, -1
}

// ContainsUserID checks if user is part of team.
//
func (t *Team) ContainsUserID(id int64) (bool, int) {

	for i, tID := range t.UserIds {
		if tID == id {
			return true, i
		}
	}
	return false, -1
}

// UpdateTeams updates an array of teams.
//
func UpdateTeams(c appengine.Context, teams []*Team) error {
	keys := make([]*datastore.Key, len(teams))
	for i := range keys {
		keys[i] = TeamKeyByID(c, teams[i].Id)
	}
	if _, err := datastore.PutMulti(c, keys, teams); err != nil {
		return err
	}
	return nil
}

// RankingByUser returns an array of user sorted by their score.
//
func (t *Team) RankingByUser(c appengine.Context, limit int) []*User {
	if limit < 0 {
		return nil
	}

	var users []*User
	var err error

	if users, err = t.Players(c); err != nil {
		return nil
	}

	sort.Sort(UserByScore(users))
	if len(users) <= limit {
		return users
	}

	return users[0:limit]
}

// TeamByAccuracy holds array of teams sorted by their accuracy
//
type TeamByAccuracy []*Team

func (a TeamByAccuracy) Len() int           { return len(a) }
func (a TeamByAccuracy) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a TeamByAccuracy) Less(i, j int) bool { return a[i].Accuracy < a[j].Accuracy }

// TournamentAcc returns the accury of a given team and for a given tournament.
//
func (t *Team) TournamentAcc(c appengine.Context, tournament *Tournament) (*Accuracy, error) {
	//query accuracy
	for _, acc := range t.AccOfTournaments {
		if acc.TournamentId == tournament.Id {
			return AccuracyByID(c, acc.AccuracyId)
		}
	}
	return nil, errors.New("model/team: accuracy not found")
}

// AddTournamentAccuracy adds accuracy to team entity and run update.
//
func (t *Team) AddTournamentAccuracy(c appengine.Context, accuracyID int64, tournamentId int64) error {

	tournamentExist := false
	for _, tid := range t.TournamentIds {
		if tid == tournamentId {
			tournamentExist = true
			break
		}
	}
	if !tournamentExist {
		return errors.New("model/team: not member of tournament")
	}
	accExist := false
	for _, acc := range t.AccOfTournaments {
		if acc.AccuracyId == accuracyID {
			accExist = true
			break
		}
	}
	if accExist {
		return errors.New("model/team: accuracy allready present")
	}

	var a AccOfTournaments
	a.AccuracyId = accuracyID
	a.TournamentId = tournamentId
	t.AccOfTournaments = append(t.AccOfTournaments, a)
	if err := t.Update(c); err != nil {
		return err
	}
	return nil
}

// UpdateAccuracy updates the global accuracy for team, with new accuracy and accuracies of other tournaments.
// From all accuracies of tournaments the team participates on we sum the overall accuracies.
// Overall accuracy is it's last element in the array of accuracies. We then normalize by the number of tournaments the teams participates on.
//
func (t *Team) UpdateAccuracy(c appengine.Context, tID int64, newAccuracy float64) error {

	sum := float64(0)
	counter := 0
	for _, tournamentAccuracy := range t.AccOfTournaments {
		if tournamentAccuracy.TournamentId == tID {
			sum += newAccuracy
			counter++
			continue
		}
		if acc, err := AccuracyByID(c, tournamentAccuracy.AccuracyId); err == nil && acc != nil {
			// only take into account tournaments with accuracies
			if len(acc.Accuracies) > 0 {
				sum += acc.Accuracies[len(acc.Accuracies)-1]
				counter++
			}
		} else if err != nil {
			log.Infof(c, "Accuracy not found %v, error:", tournamentAccuracy.AccuracyId, err)
		}
	}
	if counter > 0 {
		t.Accuracy = sum / float64(counter)
		if err := t.Update(c); err != nil {
			log.Infof(c, "Team.UpdateAccuracyAccuracy: unable to update team %v", err)
			return err
		}

		// publish new activity
		verb := fmt.Sprintf("has a new accuracy of %.2f%%", newAccuracy*100)
		t.Publish(c, "accuracy", verb, ActivityEntity{}, ActivityEntity{})
	}
	return nil
}

// Publish publishes new team activity.
//
func (t *Team) Publish(c appengine.Context, activityType string, verb string, object ActivityEntity, target ActivityEntity) error {
	var activity Activity
	activity.Type = activityType
	activity.Verb = verb
	activity.Actor = t.Entity()
	activity.Object = object
	activity.Target = target
	activity.Published = time.Now()
	activity.CreatorID = t.Id

	if err := activity.save(c); err != nil {
		return err
	}
	// add new activity id in user activity table for each member of the team
	var players []*User
	var err error
	if players, err = t.Players(c); err != nil {
		log.Errorf(c, "model/team, Publish: error occurred during Players call: %v", err)
		return nil
	}

	for _, p := range players {
		if err := activity.AddNewActivityID(c, p); err != nil {
			log.Errorf(c, "model/team, Publish: error occurred during AddNewActivityID call: %v", err)
		} else {
			if err1 := p.Update(c); err1 != nil {
				log.Errorf(c, "model/team, Publish: error occurred during update call: %v", err1)
			}
		}
	}

	return nil
}

// Entity is the Activity entity representation of a team
//
func (t *Team) Entity() ActivityEntity {
	return ActivityEntity{Id: t.Id, Type: "team", DisplayName: t.Name}
}

// AccuraciesGroupByTournament gets an array of type accuracyOverall
// which holds the accuracy information and the last 5 progression of each tournament.
//
func (t *Team) AccuraciesGroupByTournament(c appengine.Context, limit int) *[]AccuracyOverall {
	var accs []AccuracyOverall
	for _, aot := range t.AccOfTournaments {
		if acc, err := AccuracyByID(c, aot.AccuracyId); err != nil {
			log.Errorf(c, "Team.AccuraciesByTournament: Unable to retrieve accuracy entity from id, ", err)
		} else {
			var a AccuracyOverall
			a.Id = aot.AccuracyId
			a.TournamentId = aot.TournamentId
			if len(acc.Accuracies) > 0 {
				a.Accuracy = acc.Accuracies[len(acc.Accuracies)-1]
			} else {
				a.Accuracy = 1
			}
			a.Progression = make([]Progression, 0)
			counter := 0
			// get most recent accuracies with a 'limit' progression.
			for i := len(acc.Accuracies) - 1; i > -1; i-- {
				cur := acc.Accuracies[i]
				var prog Progression
				prog.Value = cur
				a.Progression = append(a.Progression, prog)
				counter++
				if counter == limit {
					break
				}
			}
			accs = append(accs, a)
		}
	}
	return &accs
}

// AccuracyByTournament gets the overall accuracy of a team in the specified tournament.
// the progression of accuracies is in reverse order to have the most reset accuracy as first element.
//
func (t *Team) AccuracyByTournament(c appengine.Context, tour *Tournament) *AccuracyOverall {
	for _, aot := range t.AccOfTournaments {
		if aot.TournamentId != tour.Id {
			continue
		}
		if acc, err := AccuracyByID(c, aot.AccuracyId); err != nil {
			log.Errorf(c, "Team.AccuraciesByTournament: Unable to retrieve accuracy entity from id, ", err)
		} else {
			var a AccuracyOverall
			a.Id = aot.AccuracyId
			a.TournamentId = aot.TournamentId
			if len(acc.Accuracies) > 0 {
				a.Accuracy = acc.Accuracies[len(acc.Accuracies)-1]
			} else {
				a.Accuracy = 1
			}
			a.Progression = make([]Progression, len(acc.Accuracies))
			for i, cur := range acc.Accuracies {
				var prog Progression
				prog.Value = cur
				a.Progression[i] = prog
			}
			return &a
		}
	}
	return nil

}

// Prices get the prices of a team.
//
func (t *Team) Prices(c appengine.Context) []*Price {
	prices := make([]*Price, len(t.PriceIds))

	for i, pid := range t.PriceIds {
		if p, err := PriceByID(c, pid); err == nil {
			prices[i] = p
		}
	}
	return prices
}

//PriceByTournament get the price of a tournament
func (t *Team) PriceByTournament(c appengine.Context, tid int64) *Price {
	for _, pid := range t.PriceIds {
		if p, err := PriceByID(c, pid); err == nil {
			if p.TournamentId == tid {
				return p
			}
		}
	}
	return nil
}
