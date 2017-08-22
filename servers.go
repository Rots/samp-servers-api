package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/pkg/errors"
	"gopkg.in/mgo.v2/bson"
)

const (
	PAGE_SIZE    = 50
	SORT_ASC     = "asc"
	SORT_DESC    = "desc"
	BY_PLAYERS   = "player"
	FILTER_PASS  = "password"
	FILTER_EMPTY = "empty"
	FILTER_FULL  = "full"
)

// Servers returns a JSON encoded array of available servers
func (app *App) Servers(w http.ResponseWriter, r *http.Request) {
	logger.Debug("getting server list")

	var (
		err    error
		page   = r.URL.Query().Get("page")
		sort   = r.URL.Query().Get("sort")
		by     = r.URL.Query().Get("by")
		filter = r.URL.Query().Get("filter")
	)

	servers, err := app.GetServers(page, sort, by, filter)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, errors.Wrap(err, "failed to get servers"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(servers)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, errors.Wrap(err, "failed to encode response"))
		return
	}
}

// GetServers returns a slice of Core objects
func (app *App) GetServers(page, sort, by, filter string) (servers []ServerCore, err error) {
	selected := []Server{}

	pageNum, err := strconv.Atoi(page)
	if err != nil {
		err = errors.Errorf("invalid 'page' argument '%s'", page)
		return
	}

	if sort != "" {
		switch sort {
		case SORT_ASC, SORT_DESC:
			break
		default:
			err = errors.Errorf("invalid 'sort' argument '%s'", sort)
			return
		}
	}

	if by != "" {
		switch by {
		case BY_PLAYERS:
			by = "core.pc"
		default:
			err = errors.Errorf("invalid 'by' argument '%s'", by)
			return
		}
	}

	var query bson.M
	if filter == "" {
		query = bson.M{}
	} else {
		switch filter {
		case FILTER_PASS:
			query = bson.M{"core.pa": false}
		case FILTER_EMPTY:
			query = bson.M{"core.pc": bson.M{"$gt": 0}}
		case FILTER_FULL:
			query = bson.M{"$where": "this.core.pc < this.core.pm"}
		default:
			err = errors.Errorf("invalid 'filter' argument '%s'", by)
			return
		}
	}

	err = app.db.Find(query).Sort(by).Skip(pageNum * PAGE_SIZE).Limit(PAGE_SIZE).All(selected)
	if err == nil {
		for i := range selected {
			servers = append(servers, selected[i].Core)
		}
	}
	return
}
