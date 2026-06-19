let tumLoglar = [];

document.addEventListener(`DOMContentLoaded`, () => {
    fetchLogs();
    setInterval(fetchLogs, 1000);
    const filterSelect = document.getElementById(`level-filter`);
    if (filterSelect) {
        filterSelect.addEventListener(`change`, ekranaBas);
    }
});

async function fetchLogs(){      //Kendime not : async olmayan bir fonksiyonda await kullanamassın.

    try {
        const response = await fetch(`http://localhost:8080/get-logs`);

        if (!response.ok) {
            throw new Error(`Sunucu hatası: ${response.status}`);
        }

        const rawData = await response.text();
        const satirlar = rawData.split(`\n`);
        const geciciLogListesi = [];

        satirlar.forEach(satir => {
            
            try {
                const logObj = JSON.parse(satir);
                geciciLogListesi.push(logObj);
            } catch (jsonErr) {
                console.warn("bozuk json satırı atlandı:", satir);
            }
        });

        tumLoglar = geciciLogListesi;
        ekranaBas();
    } catch (error) {
        console.error("Loglar çekilirken javascirpte sorun oldu:", error);
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
    container.scrollTop = container.scrollHeight;
}