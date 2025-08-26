package commands

import (
	"context"
	"fmt"

	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/matt0x6f/hashpost/internal/database"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// NewKarmaCommands returns all karma-related commands
func NewKarmaCommands(cfg *config.Config) []*cobra.Command {
	// Create the karma subcommand
	karmaCmd := &cobra.Command{
		Use:   "karma",
		Short: "Manage karma calculations and updates",
		Long:  "Commands for managing karma calculations, updates, and troubleshooting.",
	}

	// Add 'update-all' subcommand under 'karma'
	updateAllKarmaCmd := &cobra.Command{
		Use:   "update-all",
		Short: "Update karma for all pseudonyms",
		Long:  "Recalculate and update karma scores for all pseudonyms in the system. This is useful for fixing inconsistencies or after bulk operations.",
		Run: func(cmd *cobra.Command, args []string) {
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			batchSize, _ := cmd.Flags().GetInt("batch-size")
			if err := UpdateAllKarma(dryRun, batchSize, cfg); err != nil {
				log.Fatal().Err(err).Msg("Failed to update all karma")
			}
		},
	}
	updateAllKarmaCmd.Flags().Bool("dry-run", false, "Show what would be updated without making changes")
	updateAllKarmaCmd.Flags().Int("batch-size", 100, "Number of pseudonyms to process in each batch")
	karmaCmd.AddCommand(updateAllKarmaCmd)

	// Add 'update' subcommand under 'karma'
	updateKarmaCmd := &cobra.Command{
		Use:   "update [pseudonym-id]",
		Short: "Update karma for a specific pseudonym",
		Long:  "Recalculate and update karma score for a specific pseudonym by ID or display name.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			pseudonymIdentifier := args[0]
			if err := UpdateKarmaForPseudonym(pseudonymIdentifier, cfg); err != nil {
				log.Fatal().Err(err).Msg("Failed to update karma for pseudonym")
			}
		},
	}
	karmaCmd.AddCommand(updateKarmaCmd)

	// Add 'verify' subcommand under 'karma'
	verifyKarmaCmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify karma calculations",
		Long:  "Verify karma calculations for all pseudonyms and report any inconsistencies.",
		Run: func(cmd *cobra.Command, args []string) {
			if err := VerifyKarmaCalculations(cfg); err != nil {
				log.Fatal().Err(err).Msg("Failed to verify karma calculations")
			}
		},
	}
	karmaCmd.AddCommand(verifyKarmaCmd)

	return []*cobra.Command{karmaCmd}
}

// UpdateAllKarma updates karma for all pseudonyms in the system
func UpdateAllKarma(dryRun bool, batchSize int, cfg *config.Config) error {
	// Create database connection
	db, err := database.NewConnection(&cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Create DAOs
	pseudonymDAO := dao.NewPseudonymDAO(db, nil, nil, nil, nil, nil)

	log.Info().Msg("Starting karma update for all pseudonyms")
	if dryRun {
		log.Info().Msg("DRY RUN MODE - No changes will be made")
	}

	// Get total count of pseudonyms
	totalPseudonyms, err := models.Pseudonyms.Query().All(context.Background(), db)
	if err != nil {
		return fmt.Errorf("failed to get pseudonyms: %w", err)
	}

	log.Info().Int("total_pseudonyms", len(totalPseudonyms)).Int("batch_size", batchSize).Msg("Found pseudonyms to process")

	// Process in batches
	processed := 0
	updated := 0
	errors := 0

	for i := 0; i < len(totalPseudonyms); i += batchSize {
		end := i + batchSize
		if end > len(totalPseudonyms) {
			end = len(totalPseudonyms)
		}

		batch := totalPseudonyms[i:end]
		log.Info().Int("batch_start", i+1).Int("batch_end", end).Int("batch_size", len(batch)).Msg("Processing batch")

		for _, pseudonym := range batch {
			processed++
			log.Debug().Str("pseudonym_id", pseudonym.PseudonymID).Str("display_name", pseudonym.DisplayName).Msg("Processing pseudonym")

			// Calculate current karma
			currentKarma, err := pseudonymDAO.CalculateKarmaForPseudonym(context.Background(), pseudonym.PseudonymID)
			if err != nil {
				log.Error().Err(err).Str("pseudonym_id", pseudonym.PseudonymID).Msg("Failed to calculate karma")
				errors++
				continue
			}

			// Get stored karma
			storedKarma := int32(0)
			if pseudonym.KarmaScore.Valid {
				storedKarma = pseudonym.KarmaScore.V
			}

			// Check if update is needed
			if currentKarma != storedKarma {
				log.Info().
					Str("pseudonym_id", pseudonym.PseudonymID).
					Str("display_name", pseudonym.DisplayName).
					Int32("stored_karma", storedKarma).
					Int32("calculated_karma", currentKarma).
					Int32("difference", currentKarma-storedKarma).
					Msg("Karma mismatch detected")

				if !dryRun {
					// Update karma in database
					err = pseudonymDAO.UpdateKarmaForPseudonym(context.Background(), pseudonym.PseudonymID)
					if err != nil {
						log.Error().Err(err).Str("pseudonym_id", pseudonym.PseudonymID).Msg("Failed to update karma")
						errors++
						continue
					}
					log.Info().Str("pseudonym_id", pseudonym.PseudonymID).Msg("Karma updated successfully")
				}
				updated++
			} else {
				log.Debug().Str("pseudonym_id", pseudonym.PseudonymID).Msg("Karma is up to date")
			}

			// Progress indicator
			if processed%100 == 0 || processed == len(totalPseudonyms) {
				log.Info().Int("processed", processed).Int("total", len(totalPseudonyms)).Int("updated", updated).Int("errors", errors).Msg("Progress update")
			}
		}
	}

	log.Info().
		Int("total_processed", processed).
		Int("total_updated", updated).
		Int("total_errors", errors).
		Bool("dry_run", dryRun).
		Msg("Karma update completed")

	if dryRun {
		log.Info().Int("would_update", updated).Msg("DRY RUN: These pseudonyms would have been updated")
	}

	return nil
}

// UpdateKarmaForPseudonym updates karma for a specific pseudonym
func UpdateKarmaForPseudonym(pseudonymIdentifier string, cfg *config.Config) error {
	// Create database connection
	db, err := database.NewConnection(&cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Create DAOs
	pseudonymDAO := dao.NewPseudonymDAO(db, nil, nil, nil, nil, nil)

	// Try to find pseudonym by ID first, then by display name
	var pseudonym *models.Pseudonym
	var err1, err2 error

	// Try by ID
	pseudonym, err1 = pseudonymDAO.GetPseudonymByID(context.Background(), pseudonymIdentifier)
	if err1 != nil {
		log.Debug().Err(err1).Str("identifier", pseudonymIdentifier).Msg("Failed to find pseudonym by ID, trying display name")
	}

	// If not found by ID, try by display name
	if pseudonym == nil {
		pseudonym, err2 = pseudonymDAO.GetPseudonymByDisplayName(context.Background(), pseudonymIdentifier)
		if err2 != nil {
			return fmt.Errorf("failed to find pseudonym by ID '%s' or display name '%s': %w", pseudonymIdentifier, pseudonymIdentifier, err2)
		}
	}

	if pseudonym == nil {
		return fmt.Errorf("pseudonym not found with identifier: %s", pseudonymIdentifier)
	}

	log.Info().
		Str("pseudonym_id", pseudonym.PseudonymID).
		Str("display_name", pseudonym.DisplayName).
		Msg("Found pseudonym, calculating karma")

	// Get current stored karma
	storedKarma := int32(0)
	if pseudonym.KarmaScore.Valid {
		storedKarma = pseudonym.KarmaScore.V
	}

	// Calculate new karma
	newKarma, err := pseudonymDAO.CalculateKarmaForPseudonym(context.Background(), pseudonym.PseudonymID)
	if err != nil {
		return fmt.Errorf("failed to calculate karma: %w", err)
	}

	log.Info().
		Str("pseudonym_id", pseudonym.PseudonymID).
		Str("display_name", pseudonym.DisplayName).
		Int32("stored_karma", storedKarma).
		Int32("calculated_karma", newKarma).
		Int32("difference", newKarma-storedKarma).
		Msg("Karma calculation completed")

	if newKarma == storedKarma {
		log.Info().Msg("Karma is already up to date")
		return nil
	}

	// Update karma in database
	err = pseudonymDAO.UpdateKarmaForPseudonym(context.Background(), pseudonym.PseudonymID)
	if err != nil {
		return fmt.Errorf("failed to update karma: %w", err)
	}

	log.Info().
		Str("pseudonym_id", pseudonym.PseudonymID).
		Str("display_name", pseudonym.DisplayName).
		Int32("old_karma", storedKarma).
		Int32("new_karma", newKarma).
		Msg("Karma updated successfully")

	return nil
}

// VerifyKarmaCalculations verifies karma calculations for all pseudonyms
func VerifyKarmaCalculations(cfg *config.Config) error {
	// Create database connection
	db, err := database.NewConnection(&cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Create DAOs
	pseudonymDAO := dao.NewPseudonymDAO(db, nil, nil, nil, nil, nil)

	log.Info().Msg("Starting karma verification for all pseudonyms")

	// Get all pseudonyms
	pseudonyms, err := models.Pseudonyms.Query().All(context.Background(), db)
	if err != nil {
		return fmt.Errorf("failed to get pseudonyms: %w", err)
	}

	log.Info().Int("total_pseudonyms", len(pseudonyms)).Msg("Found pseudonyms to verify")

	// Verify each pseudonym
	verified := 0
	mismatches := 0
	errors := 0

	for _, pseudonym := range pseudonyms {
		verified++

		// Get stored karma
		storedKarma := int32(0)
		if pseudonym.KarmaScore.Valid {
			storedKarma = pseudonym.KarmaScore.V
		}

		// Calculate current karma
		calculatedKarma, err := pseudonymDAO.CalculateKarmaForPseudonym(context.Background(), pseudonym.PseudonymID)
		if err != nil {
			log.Error().Err(err).Str("pseudonym_id", pseudonym.PseudonymID).Msg("Failed to calculate karma")
			errors++
			continue
		}

		// Check for mismatch
		if calculatedKarma != storedKarma {
			log.Warn().
				Str("pseudonym_id", pseudonym.PseudonymID).
				Str("display_name", pseudonym.DisplayName).
				Int32("stored_karma", storedKarma).
				Int32("calculated_karma", calculatedKarma).
				Int32("difference", calculatedKarma-storedKarma).
				Msg("Karma mismatch detected")
			mismatches++
		}

		// Progress indicator
		if verified%100 == 0 || verified == len(pseudonyms) {
			log.Info().Int("verified", verified).Int("total", len(pseudonyms)).Int("mismatches", mismatches).Int("errors", errors).Msg("Verification progress")
		}
	}

	log.Info().
		Int("total_verified", verified).
		Int("total_mismatches", mismatches).
		Int("total_errors", errors).
		Msg("Karma verification completed")

	if mismatches > 0 {
		log.Warn().Int("mismatches", mismatches).Msg("Karma mismatches detected - consider running 'karma update-all' to fix")
	} else {
		log.Info().Msg("All karma calculations are consistent")
	}

	return nil
}
