package database

import (
	"database/sql"

	"github.com/nsec/askgod/api"
)

// GetTimeline generates the current timeline
func (db *DB) GetTimeline(team *api.AdminTeam) ([]api.TimelineEntry, error) {
	// Return a list of score entries
	resp := []api.TimelineEntry{}

	// Query all the scores from the database
	var rows *sql.Rows
	var err error

	if team == nil {
		rows, err = db.Query("SELECT team.id, team.country, team.name, team.website, score.value, score.submit_time FROM score LEFT JOIN team ON team.id=score.teamid ORDER BY team.id ASC, score.submit_time ASC;")
		if err != nil {
			return nil, err
		}
	} else {
		rows, err = db.Query("SELECT team.id, team.country, team.name, team.website, score.value, score.submit_time FROM score LEFT JOIN team ON team.id=score.teamid WHERE team.id=$1 ORDER BY team.id ASC, score.submit_time ASC;", team.ID)
		if err != nil {
			return nil, err
		}
	}
	defer rows.Close()

	// Iterate through the results
	total := int64(0)
	entry := api.TimelineEntry{}
	for rows.Next() {
		rowTeam := api.Team{}
		rowScore := api.TimelineEntryScore{}

		err := rows.Scan(&rowTeam.ID, &rowTeam.Country, &rowTeam.Name, &rowTeam.Website, &rowScore.Value, &rowScore.SubmitTime)
		if err != nil {
			return nil, err
		}

		if entry.Team.ID != rowTeam.ID {
			if entry.Team.ID > 0 {
				resp = append(resp, entry)
			}

			entry = api.TimelineEntry{
				Team:  rowTeam,
				Score: []api.TimelineEntryScore{},
			}
			total = 0
		}

		total += rowScore.Value
		rowScore.Total = total

		entry.Score = append(entry.Score, rowScore)
	}

	if entry.Team.ID > 0 {
		resp = append(resp, entry)
	}

	// Check for any error that might have happened
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return resp, nil
}
