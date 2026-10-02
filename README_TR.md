# FlowPilot

**İş akışlarınızı otomatikleştirin. Altyapınız size ait kalsın.**

FlowPilot, PostgreSQL üzerinde kalıcı workflow sürümleri ve execution kayıtları tutan, self-hosted otomasyon platformu çekirdeğidir.

## Hızlı başlangıç

1. `.env.example` dosyasını `.env` olarak kopyalayın ve sırları değiştirin.
2. Docker Desktop’ı çalıştırın.
3. `docker compose up --build` komutunu kullanın.
4. `http://localhost:3000` adresini açın.

Bir workflow draft’ı kaydedin, publish edin ve manual execution başlatın. Webhook trigger kullanıyorsanız trigger yapılandırmasına benzersiz bir `webhookId` ekleyin; ardından `POST /hooks/{webhookId}` ile çağırın.

Bu sürümün üretimde kullanılmadan önce tamamlanması gereken güvenlik özellikleri vardır. Ayrıntı için [SECURITY.md](docs/SECURITY.md) dosyasına bakın.
