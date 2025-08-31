
function aggiornaDatiTemperatura(chart) {
    // Assumo un endpoint che restituisca: { "current": 22.5, "average": 21.8, "max": 25.1, "min": 19.5 }
    fetch("http://localhost:8080/api/temperature-stats")
        .then(r => {
            if (!r.ok) {
                throw new Error('Network response was not ok');
            }
            return r.json();
        })
        .then(stats => {
           
            document.getElementById('temp-current').textContent = stats.current.toFixed(1);
            document.getElementById('temp-avg').textContent = stats.average.toFixed(1);
            document.getElementById('temp-max').textContent = stats.max.toFixed(1);
            document.getElementById('temp-min').textContent = stats.min.toFixed(1);

            // 2. Aggiorna il grafico con la temperatura attuale
            const data = chart.data;
            data.datasets[0].data.shift();
            data.datasets[0].data.push(stats.current);

            const lastLabel = data.labels[data.labels.length - 1];
            const newLabelIndex = parseInt(lastLabel.substring(1)) + 1;
            data.labels.shift();
            data.labels.push('T' + newLabelIndex);

            chart.update();
        })
        .catch(error => {
            console.error("Errore nell'aggiornare i dati di temperatura:", error);
            document.getElementById('temp-current').textContent = '--';
            document.getElementById('temp-avg').textContent = '--';
            document.getElementById('temp-max').textContent = '--';
            document.getElementById('temp-min').textContent = '--';
        });
}

function controlloAllarmi(){
    fetch("http://localhost:8080/api/get-alarms")
        .then(r => {
            if (!r.ok) {
                throw new Error('Network response was not ok');
            }
            return r.json();
        })
        .then(alarms => {
            const button = document.getElementById('reset-alarm')
            button.classList.toggle('in-alarm', alarms.attivo);
            if (alarms.attivo){
                button.textContent = "Reset Allarme";
            } else {
                button.textContent = "No Allarmi";   
            }
        })
        .catch(error => {
            console.error("Errore nell'aggiornare lo stato degli allarmi ", error);
        });
}


function controlloModalita(){
    fetch("http://localhost:8080/api/get-operative-mode")
        .then(r => {
            if (!r.ok) {
                throw new Error('Network response was not ok');
            }
            return r.json();
        })
        .then(data => {
            const text = document.getElementById('actual-mode')
            const button = document.getElementById('cambia-modalita')
            if (data.manuale){
                text.textContent = "Modalità Attuale: Manuale";
                button.textContent = "Modalità Automatica"
            } else {
                text.textContent = "Modalità Attuale: Automatico";
                button.textContent = "Modalità Manuale"
            }
        })
        .catch(error => {
            console.error("Errore nell'aggiornare lo stato degli allarmi ", error);
        });
}

function aggiornaStatoDispositivi() {
    fetch("http://localhost:8080/api/devices-states")
        .then(r => {
            if (!r.ok){
                throw new Error('Network response was not ok');
            }
            return r.json()
        })
        .then(stato => {
            document.querySelectorAll('[data-device]').forEach(link => {
                const nome = link.getAttribute('data-device');
                const isOnline = stato[nome] === true;
                link.classList.toggle('online', isOnline);
                link.classList.toggle('offline', !isOnline);
            });
        })
        .catch(() => {
            document.querySelectorAll('[data-device]').forEach(link => {
                link.classList.add('offline');
                link.classList.remove('online');
            });
        });
}

function createChartConfig(data) {
    return {
        type: 'line',
        data: data,
        options: {
            responsive: true,
            animation: false,
            plugins: {
                legend: {
                    display: false
                }
            }
        }
    };
}


function sendPostRequest(url) {
    fetch(url, {
            method: 'POST'
        })
        .then(response => {
            if (!response.ok) {
                throw new Error(`Network response was not ok for ${url}`);
            }
            console.log(`POST a ${url} riuscito.`);
        })
        .catch(error => {
            console.error(`Errore nella richiesta POST a ${url}:`, error);
        });
}

function aggiornaStatoSistema() {
    fetch("http://localhost:8080/api/system-status")
        .then(response => {
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            return response.json();
        })
        .then(data => {
            document.getElementById('system-status').textContent = data.status;
        })
        .catch(error => {
            console.error("Errore nell'aggiornare lo stato del sistema:", error);
            document.getElementById('system-status').textContent = 'ERRORE';
        });
}


function aggiornaPosizioneFinestra() {
    fetch("http://localhost:8080/api/window-position")
        .then(response => {
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            return response.json();
        })
        .then(data => {
            document.getElementById('window-level').textContent = data.position;
        })
        .catch(error => {
            console.error("Errore nell'aggiornare la posizione della finestra:", error);
            document.getElementById('window-level').textContent = '--';
        });
}



document.addEventListener('DOMContentLoaded', () => {
    // Inizializzazione Grafico
    const NUM_PUNTI = 100;
    const initialLabels = Array.from({
        length: NUM_PUNTI
    }, (_, i) => `T${i + 1}`);
    const initialData = {
        labels: initialLabels,
        datasets: [{
            label: 'Temperatura (°C)',
            data: Array.from({
                length: NUM_PUNTI
            }, () => 0),
            borderColor: 'rgb(75, 192, 192)',
            tension: 0.1
        }]
    };

    const chartConfig = createChartConfig(initialData);
    const tempChart = new Chart(document.getElementById('tempChart'), chartConfig);

    aggiornaDatiTemperatura(tempChart);
    setInterval(() => aggiornaDatiTemperatura(tempChart), 100);

    aggiornaStatoDispositivi();
    setInterval(aggiornaStatoDispositivi, 500);

    aggiornaStatoSistema();
    setInterval(aggiornaStatoSistema, 500);

    aggiornaPosizioneFinestra();
    setInterval(aggiornaPosizioneFinestra, 500);

    controlloAllarmi();
    setInterval(controlloAllarmi, 1000);

    controlloModalita();
    setInterval(controlloModalita, 1000);

    document.getElementById('cambia-modalita').addEventListener('click', () => {
        sendPostRequest('http://localhost:8080/api/change-mode');
    });

    document.getElementById('open-window').addEventListener('click', () => {
        sendPostRequest('http://localhost:8080/api/open-window');
    });

    document.getElementById('close-window').addEventListener('click', () => {
        sendPostRequest('http://localhost:8080/api/close-window');
    });

    document.getElementById('reset-alarm').addEventListener('click', () => {
        sendPostRequest('http://localhost:8080/api/reset-alarm');
    });
});