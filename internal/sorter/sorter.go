package sorter

import (
	"cmp"
	"slices"

	c "github.com/achannarasappa/ticker/v5/internal/common"
)

// Sorter represents a function that sorts quotes
type Sorter func([]*c.Asset) []*c.Asset

// NewSorter creates a sorting function
func NewSorter(sort string) Sorter {
	var sortDict = map[string]Sorter{
		"alpha":  sortByAlpha,
		"symbol": sortByAlpha,
		"value":  sortByValue,
		"user":   sortByUser,
		"close":  sortByClose,
		"volume":     sortByVolume,
		"mktcap":     sortByMarketCap,
		"market_cap": sortByMarketCap,
		"high":       sortByHigh,
		"low":        sortByLow,
		"open":       sortByOpen,
		"change":     sortByChange,
	}
	if sorter, ok := sortDict[sort]; ok {
		return sorter
	}

	return sortByChange
}

func sortByUser(assets []*c.Asset) []*c.Asset {

	assetCount := len(assets)

	if assetCount <= 0 {
		return assets
	}

	slices.SortStableFunc(assets, func(a, b *c.Asset) int {
		return cmp.Compare(a.Meta.OrderIndex, b.Meta.OrderIndex)
	})

	return assets

}

func sortByAlpha(assetsIn []*c.Asset) []*c.Asset {

	assetCount := len(assetsIn)

	if assetCount <= 0 {
		return assetsIn
	}

	assets := make([]*c.Asset, assetCount)
	copy(assets, assetsIn)

	slices.SortStableFunc(assets, func(a, b *c.Asset) int {
		return cmp.Compare(a.Symbol, b.Symbol)
	})

	return assets
}

func sortByValue(assetsIn []*c.Asset) []*c.Asset {

	assetCount := len(assetsIn)

	if assetCount <= 0 {
		return assetsIn
	}

	assets := make([]*c.Asset, assetCount)
	copy(assets, assetsIn)

	activeAssets, inactiveAssets := splitActiveAssets(assets)

	slices.SortStableFunc(inactiveAssets, func(a, b *c.Asset) int {
		return cmp.Compare(b.Position.Value, a.Position.Value)
	})

	slices.SortStableFunc(activeAssets, func(a, b *c.Asset) int {
		return cmp.Compare(b.Position.Value, a.Position.Value)
	})

	return append(activeAssets, inactiveAssets...)
}

func sortByChange(assetsIn []*c.Asset) []*c.Asset {

	assetCount := len(assetsIn)

	if assetCount <= 0 {
		return assetsIn
	}

	assets := make([]*c.Asset, assetCount)
	copy(assets, assetsIn)

	activeAssets, inactiveAssets := splitActiveAssets(assets)

	slices.SortStableFunc(activeAssets, func(a, b *c.Asset) int {
		return cmp.Compare(b.QuotePrice.ChangePercent, a.QuotePrice.ChangePercent)
	})

	slices.SortStableFunc(inactiveAssets, func(a, b *c.Asset) int {
		return cmp.Compare(b.QuotePrice.ChangePercent, a.QuotePrice.ChangePercent)
	})

	return append(activeAssets, inactiveAssets...)

}

func sortByClose(assetsIn []*c.Asset) []*c.Asset {
	assetCount := len(assetsIn)
	if assetCount <= 0 {
		return assetsIn
	}
	assets := make([]*c.Asset, assetCount)
	copy(assets, assetsIn)

	slices.SortStableFunc(assets, func(a, b *c.Asset) int {
		aClose := a.QuotePrice.PriceRegularClose
		if aClose == 0.0 {
			aClose = a.QuotePrice.Price
		}
		bClose := b.QuotePrice.PriceRegularClose
		if bClose == 0.0 {
			bClose = b.QuotePrice.Price
		}
		return cmp.Compare(bClose, aClose)
	})

	return assets
}

func sortByVolume(assetsIn []*c.Asset) []*c.Asset {
	assetCount := len(assetsIn)
	if assetCount <= 0 {
		return assetsIn
	}
	assets := make([]*c.Asset, assetCount)
	copy(assets, assetsIn)

	slices.SortStableFunc(assets, func(a, b *c.Asset) int {
		return cmp.Compare(b.QuoteExtended.Volume, a.QuoteExtended.Volume)
	})

	return assets
}

func sortByMarketCap(assetsIn []*c.Asset) []*c.Asset {
	assetCount := len(assetsIn)
	if assetCount <= 0 {
		return assetsIn
	}
	assets := make([]*c.Asset, assetCount)
	copy(assets, assetsIn)

	slices.SortStableFunc(assets, func(a, b *c.Asset) int {
		return cmp.Compare(b.QuoteExtended.MarketCap, a.QuoteExtended.MarketCap)
	})

	return assets
}

func sortByHigh(assetsIn []*c.Asset) []*c.Asset {
	assetCount := len(assetsIn)
	if assetCount <= 0 {
		return assetsIn
	}
	assets := make([]*c.Asset, assetCount)
	copy(assets, assetsIn)

	slices.SortStableFunc(assets, func(a, b *c.Asset) int {
		return cmp.Compare(b.QuotePrice.PriceDayHigh, a.QuotePrice.PriceDayHigh)
	})

	return assets
}

func sortByLow(assetsIn []*c.Asset) []*c.Asset {
	assetCount := len(assetsIn)
	if assetCount <= 0 {
		return assetsIn
	}
	assets := make([]*c.Asset, assetCount)
	copy(assets, assetsIn)

	slices.SortStableFunc(assets, func(a, b *c.Asset) int {
		return cmp.Compare(a.QuotePrice.PriceDayLow, b.QuotePrice.PriceDayLow)
	})

	return assets
}

func sortByOpen(assetsIn []*c.Asset) []*c.Asset {
	assetCount := len(assetsIn)
	if assetCount <= 0 {
		return assetsIn
	}
	assets := make([]*c.Asset, assetCount)
	copy(assets, assetsIn)

	slices.SortStableFunc(assets, func(a, b *c.Asset) int {
		return cmp.Compare(b.QuotePrice.PriceOpen, a.QuotePrice.PriceOpen)
	})

	return assets
}

func splitActiveAssets(assets []*c.Asset) ([]*c.Asset, []*c.Asset) {

	activeAssets := make([]*c.Asset, 0)
	inactiveAssets := make([]*c.Asset, 0)

	for _, asset := range assets {
		if asset.Exchange.IsActive {
			activeAssets = append(activeAssets, asset)
		} else {
			inactiveAssets = append(inactiveAssets, asset)
		}
	}

	return activeAssets, inactiveAssets
}
