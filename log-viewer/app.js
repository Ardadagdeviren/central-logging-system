let tumLoglar = [];
const MAX_LOG_SAYISI = 50;

document.addEventListener(`DOMContentLoaded`, () => {
    fetchLogs();

    const filterSelect = document.getElementById(`level-filter`);
    if (filterSelect) {
        filterSelect.addEventListener(`change`, ekranaBas);
    }
});

async function fetchLogs() {      //Kendime not : async olmayan bir fonksiyonda await kullanamassın.

    try {
        const response = await fetch(`http://localhost:8080/get-logs`);

        if (!response.ok) {
            throw new Error(`Sunucu hatası: ${response.status}`);
        }

        const rawData = await response.text();
        const satirlar = rawData.split(`\n`);      //split geriye dizi döndürüyor
        const geciciLogListesi = [];

        satirlar.forEach(satir => {

            try {
                const logObj = JSON.parse(satir);      //json formatını js objesine dönüştürür
                geciciLogListesi.push(logObj);
            } catch (jsonErr) {         //jsonErr catchin içinde bir parametre otomatik js tanımlar
                console.warn("bozuk json satırı atlandı:", satir);
            }
        });

        tumLoglar = geciciLogListesi.slice(-MAX_LOG_SAYISI);
        ekranaBas();
    } catch (error) {
        console.error("Loglar çekilirken javascirpte sorun oldu:", error);
    }
    finally {

        setTimeout(fetchLogs, 7000);
    }
}

function ekranaBas() {
    const container = document.getElementById(`log-container`);
    const filterSelect = document.getElementById(`level-filter`);



    const secilenFiltre = filterSelect.value;

    container.innerHTML = ``;
    tumLoglar.forEach(yeniLog => {
        if (secilenFiltre === `ALL` || yeniLog.level === secilenFiltre) {
            renderLogLine(yeniLog, container);
        }
    });
}



function renderLogLine(yeniLog, container) {
    const logSatiri = document.createElement(`div`);
    logSatiri.className = `log-line`;

    if (yeniLog.level === `ERROR`) {
        logSatiri.classList.add(`log-error`);
    } else if (yeniLog.level === `WARN`) {
        logSatiri.classList.add(`log-warn`);
    } else if (yeniLog.level === `INFO`) {
        logSatiri.classList.add(`log-info`);
    }

    logSatiri.innerText = `[${yeniLog.timestamp}] [${yeniLog.service}] [${yeniLog.level}]: ${yeniLog.message}`;

    container.appendChild(logSatiri);

}