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
			c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"sync_committee": syncCommittee})
	}
}
