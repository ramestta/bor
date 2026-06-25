// Ramestta file-based address blocklist (ported from bor 909a993 core/tx_pool.go).
// Reads /etc/bor/blocklist.txt, auto-reloads every 60s, plus hardcoded fallbacks.
// Hook lives in LegacyPool.validateTx (legacypool.go): rejects tx from/to a blocked address.
package legacypool

import (
	"bufio"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

const blocklistFilePath = "/etc/bor/blocklist.txt"

var (
	blocklistMu         sync.RWMutex
	blockedAddressesMap = make(map[common.Address]bool)
	blocklistLoaded     bool
)

func init() {
	loadBlocklist()
	go blocklistReloader()
}

// loadBlocklist reads addresses from the blocklist file (hardcoded fallbacks always included).
func loadBlocklist() {
	newMap := make(map[common.Address]bool)
	hardcoded := []string{
		"0x77705dCCBd18318B5726753faCdfB35DA9Ee3D94",
		"0xD67105cd5faE05f71a37feaBCb51e381B1Cfa6F2",
		"0xD13a76AdB25A2Bb9EE81A0C8828Df8660f70f4c9",
		"0x89F44d3455f7EF111E34d9bdb104F9489E6dCD61",
		"0x6f1F0dab7f39E55C0D27a4823d402E88a847c4B7",
	}
	for _, addr := range hardcoded {
		newMap[common.HexToAddress(addr)] = true
	}
	file, err := os.Open(blocklistFilePath)
	if err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if idx := strings.Index(line, "#"); idx > 0 {
				line = strings.TrimSpace(line[:idx])
			}
			if common.IsHexAddress(line) {
				newMap[common.HexToAddress(line)] = true
			}
		}
	}
	blocklistMu.Lock()
	blockedAddressesMap = newMap
	blocklistLoaded = true
	blocklistMu.Unlock()
}

// blocklistReloader periodically reloads the blocklist file (no bor restart needed).
func blocklistReloader() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		loadBlocklist()
	}
}

// isAddressBlocked reports whether addr is in the blocklist.
func isAddressBlocked(addr common.Address) bool {
	blocklistMu.RLock()
	defer blocklistMu.RUnlock()
	return blockedAddressesMap[addr]
}
