package main

import (
	"log"
	"time"

	"github.com/spendly/backend/internal/bootstrap"
	"github.com/spendly/backend/internal/config"
	"github.com/spendly/backend/internal/domain"
)

func main() {
	cfg, _ := config.LoadConfig()
	db, err := bootstrap.InitDatabase(cfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Memasukkan data dummy langsung ke DB...")

	// 1. Buat Kategori
	catFood := domain.Category{
		Label: "Makanan",
		Icon:  "fastfood",
		Color: "#FF5733",
		Type:  "expense",
	}
	catSalary := domain.Category{
		Label: "Gaji",
		Icon:  "payments",
		Color: "#2ECC71",
		Type:  "income",
	}
	
	db.FirstOrCreate(&catFood, domain.Category{Label: "Makanan"})
	db.FirstOrCreate(&catSalary, domain.Category{Label: "Gaji"})

	// 2. Buat Transaksi
	tx1 := domain.Transaction{
		Title:      "Gaji Bulan Ini",
		Amount:     5000000,
		Date:       time.Now(),
		CategoryID: catSalary.ID,
		Type:       "income",
		Note:       "Bonus performa",
	}
	tx2 := domain.Transaction{
		Title:      "Makan Siang Bakso",
		Amount:     25000,
		Date:       time.Now(),
		CategoryID: catFood.ID,
		Type:       "expense",
		Note:       "Lapar sekali",
	}

	db.Create(&tx1)
	db.Create(&tx2)

	log.Println("Berhasil! Data kategori dan transaksi sudah masuk.")
}
