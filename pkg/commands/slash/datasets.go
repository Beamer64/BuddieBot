package slash

import (
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/Beamer64/BuddieBot/pkg/config"
)

// wyrPolls holds the parsed Would You Rather poll data. Populated once at
// startup via LoadStaticData and read-only afterwards — safe to read from
// any goroutine without locking.
var wyrPolls []wyrPoll

// LoadStaticData loads every static dataset the slash package depends on.
// Call this once during bot startup (from bot.Init). Returns an error if any
// required dataset is missing or malformed — the bot should refuse to start
// in that case so broken data is caught immediately, not on first invocation.
func LoadStaticData(cfg *config.Configs) error {
	if err := loadWYR(cfg); err != nil {
		return fmt.Errorf("loading WYR data: %w", err)
	}
	return nil
}

func loadWYR(cfg *config.Configs) error {
	path := filepath.Join(cfg.Configs.ReqFileDirs.Datasets, "WYR.csv")

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return err
	}

	polls, err := parseWYRRecords(records)
	if err != nil {
		return err
	}

	wyrPolls = polls
	log.Printf("Loaded %d WYR polls", len(polls))
	return nil
}

// parseWYRRecords converts CSV records (header + rows) into WYR polls.
// Malformed rows (too few columns, non-numeric vote counts) are logged and
// skipped rather than aborting the whole load. Returns an error only when no
// usable rows remain.
//
// Pure function — no I/O, suitable for unit tests.
func parseWYRRecords(records [][]string) ([]wyrPoll, error) {
	if len(records) <= 1 {
		return nil, errors.New("no data rows found in WYR CSV")
	}

	polls := make([]wyrPoll, 0, len(records)-1)
	for rowNum, row := range records[1:] {
		if len(row) < 4 {
			log.Printf("WYR CSV row %d skipped (only %d columns)", rowNum+2, len(row))
			continue
		}
		votesA, errA := strconv.Atoi(row[1])
		votesB, errB := strconv.Atoi(row[3])
		if errA != nil || errB != nil {
			log.Printf("WYR CSV row %d skipped (non-numeric vote: %v / %v)", rowNum+2, errA, errB)
			continue
		}
		polls = append(polls, wyrPoll{
			OptionA: row[0],
			VotesA:  votesA,
			OptionB: row[2],
			VotesB:  votesB,
		})
	}

	if len(polls) == 0 {
		return nil, errors.New("no valid WYR rows after parsing")
	}
	return polls, nil
}
