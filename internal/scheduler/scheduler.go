package scheduler

import (
	"fmt"
	"ngabaca/database"
	"ngabaca/internal/model"
	"time"

	"gorm.io/gorm"
)

// CancelExpiredOrders adalah fungsi yang akan dijalankan oleh cron job.
// Fungsi ini mencari pembayaran yang 'pending' dan sudah melewati waktu kedaluwarsa.
func CancelExpiredOrders() {
	fmt.Printf("[%s] Menjalankan tugas pembatalan pesanan kedaluwarsa...\n", time.Now().Format("2006-01-02 15:04:05"))

	var expiredPayments []model.Payment

	// 1. Cari semua pembayaran yang statusnya 'pending' dan sudah kedaluwarsa.
	err := database.DB.Where("status = ? AND expires_at < ?", "pending", time.Now()).Find(&expiredPayments).Error
	if err != nil {
		fmt.Println("Error saat mencari pembayaran kedaluwarsa:", err)
		return
	}

	if len(expiredPayments) == 0 {
		fmt.Println("Tidak ada pesanan kedaluwarsa yang ditemukan.")
		return
	}

	fmt.Printf("Ditemukan %d pesanan kedaluwarsa. Memproses pembatalan...\n", len(expiredPayments))

	// 2. Gunakan transaksi untuk memastikan semua operasi (batal & kembalikan stok) berhasil.
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		for _, payment := range expiredPayments {
			// Update status payment menjadi 'failed'
			payment.Status = "failed"
			if err := tx.Save(&payment).Error; err != nil {
				return err
			}

			// Update status order menjadi 'batal'
			if err := tx.Model(&model.Order{}).Where("id = ?", payment.OrderID).Update("status", "batal").Error; err != nil {
				return err
			}

			// 3. Kembalikan stok buku
			var items []model.OrderItem
			if err := tx.Where("order_id = ?", payment.OrderID).Find(&items).Error; err != nil {
				return err
			}

			for _, item := range items {
				// Gunakan gorm.Expr untuk operasi atomik (menghindari race condition)
				err := tx.Model(&model.Book{}).Where("id = ?", item.BookID).
					Update("stock", gorm.Expr("stock + ?", item.Quantity)).Error

				if err != nil {
					return err
				}
				fmt.Printf("  - Stok untuk buku ID %d dikembalikan sebanyak %d\n", item.BookID, item.Quantity)
			}
		}
		return nil
	})

	if err != nil {
		fmt.Println("Error saat memproses transaksi pembatalan:", err)
	} else {
		fmt.Printf("Berhasil membatalkan %d pesanan kedaluwarsa.\n", len(expiredPayments))
	}
}
