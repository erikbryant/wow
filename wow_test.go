package main

import (
	"github.com/erikbryant/database"
	"testing"
)

var (
	auction1 = database.Auction{
		Auc:      1,
		Item:     100,
		Bid:      100,
		Buyout:   110,
		Quantity: 1,
	}
	auction2 = database.Auction{
		Auc:      2,
		Item:     200,
		Bid:      100,
		Buyout:   11000,
		Quantity: 1,
	}
	auction3 = database.Auction{
		Auc:      3,
		Item:     300,
		Bid:      100,
		Buyout:   11000,
		Quantity: 100,
	}
)

func TestBargains(t *testing.T) {
	testCases := []struct {
		auctions    map[int64]database.Auction
		goods       map[int64]int64
		expectedBid []int64
		expectedBuy []int64
	}{
		{
			// Auctions.
			map[int64]database.Auction{
				1: auction1,
				2: auction2,
				3: auction3,
			},
			// Goods to look for bargains on.
			map[int64]int64{
				100: 1000,
				200: 2000,
				300: 3000,
			},
			// Expected bids.
			[]int64{
				2,
			},
			// Expected buys (if a buy is signaled, it will not
			// also signal a bid).
			[]int64{
				1,
				3,
			},
		},
	}

	for test, testCase := range testCases {
		answerBid, answerBuy := bargains(testCase.auctions, testCase.goods)
		if len(answerBid) != len(testCase.expectedBid) {
			t.Errorf("ERROR: For test %d, expected bids %v, got %v", test, testCase.expectedBid, answerBid)
		}
		for i := range answerBid {
			if answerBid[i] != testCase.expectedBid[i] {
				t.Errorf("ERROR: For test %d, expected bids %v, got %v", test, testCase.expectedBid, answerBid)
			}
		}

		if len(answerBuy) != len(testCase.expectedBuy) {
			t.Errorf("ERROR: For test %d, expected buys %v, got %v", test, testCase.expectedBuy, answerBuy)
		}
		for i := range answerBuy {
			if answerBuy[i] != testCase.expectedBuy[i] {
				t.Errorf("ERROR: For test %d, expected buys %v, got %v", test, testCase.expectedBuy, answerBuy)
			}
		}
	}
}

func TestCoinsToString(t *testing.T) {
	testCases := []struct {
		amount   int64
		expected string
	}{
		{0, "0"},
		{-1, "-1"},
		{100, "1.00"},
		{-100, "-1.00"},
		{10000, "1.00.00"},
		{-10000, "-1.00.00"},
		{123456, "12.34.56"},
	}

	for _, testCase := range testCases {
		answer := coinsToString(testCase.amount)
		if answer != testCase.expected {
			t.Errorf("ERROR: For %d expected '%s', got '%s'", testCase.amount, testCase.expected, answer)
		}
	}
}
