package handlers

import (
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/s6thgehr/sf-challenge/api"
)

func SyncDutiesHandler(client *ethclient.Client, rpc string) gin.HandlerFunc {
	return func(c *gin.Context) {
		slot, err := strconv.Atoi(c.Param("slot"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
			return
		}

		syncCommittee, err := api.FetchSyncCommittee(rpc, slot)
		if err != nil {
			if err.Error() == "400" {
				c.JSON(http.StatusBadRequest, gin.H{"message": "requested slot is too far in the future to have duties available"})
				return
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
				return
			}
		}

		var channel = make(chan string, 512)
		var syncCommitteeAddresses []string

		go api.ConvertValidatorIndexToPubkey(rpc, syncCommittee, slot, channel)
		for address := range channel {
			syncCommitteeAddresses = append(syncCommitteeAddresses, address)
		}

		c.JSON(http.StatusOK, gin.H{"sync_committee": syncCommitteeAddresses})

	}
}
