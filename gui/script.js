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

function aggiornaDatiTemperatura(data, chart) {
    document.getElementById('temp-current').textContent = data.CurrentTemp.toFixed(1);
    document.getElementById('temp-avg').textContent = data.AverageTemp.toFixed(1);
    document.getElementById('temp-max').textContent = data.MaxTemp.toFixed(1);
    document.getElementById('temp-min').textContent = data.MinTemp.toFixed(1);

    // Aggiorna il grafico
    const chartData = chart.data;
    chartData.datasets[0].data.shift();
    chartData.datasets[0].data.push(data.CurrentTemp);

    const lastLabel = chartData.labels[chartData.labels.length - 1];
    const newLabelIndex = parseInt(lastLabel.substring(1)) + 1;
    chartData.labels.shift();
    chartData.labels.push('T' + newLabelIndex);

    chart.update();
}

function aggiornaStatoDispositivi(devicesStatus) {
    document.querySelectorAll('[data-device]').forEach(link => {
        const nome = link.getAttribute('data-device');
        const isOnline = devicesStatus[nome] === true;
        link.classList.toggle('online', isOnline);
        link.classList.toggle('offline', !isOnline);
    });
}

function aggiornaStatoSistema(statusString) {
    document.getElementById('system-status').textContent = statusString;
}

function aggiornaPosizioneFinestra(position) {
    document.getElementById('window-level').textContent = (position * (100/90)).toFixed(2) ;
}

function aggiornaAllarmi(status) {
    const button = document.getElementById('reset-alarm');
    var inAlarm = status.trim ().toUpperCase() === "ALARM"
    console.log("Stato allarme:", status, "inAlarm:", inAlarm);
    if (inAlarm) {
        button.classList.add('in-alarm');
        button.textContent = "Reset Allarme";
      } else {
        button.classList.remove('in-alarm');
        button.textContent = "No Allarmi";
    }
}

function aggiornaModalita(mode) {
    console.log("Modalità ricevuta:", mode); // <--- aggiungi questo
    const text = document.getElementById('actual-mode');
    const button = document.getElementById('cambia-modalita');
    if (mode.trim().toUpperCase() === "MANUALE") {
        text.textContent = "Modalità Attuale: Manuale";
        button.textContent = "Modalità Automatica";
    } else {
        text.textContent = "Modalità Attuale: Automatico";
        button.textContent = "Modalità Manuale";
    }
}

function update(chart) {
 console.log("ciao")
    fetch("http://localhost:8080/api/system-status")
        .then(response => {
            if (!response.ok) throw new Error('Network response was not ok');
            return response.json();
        })
        .then(data => {
           
            aggiornaDatiTemperatura(data, chart);
            aggiornaStatoDispositivi(data.DevicesOnline);
            aggiornaStatoSistema(data.StatusString);
            aggiornaPosizioneFinestra(data.WindowPosition);
            aggiornaAllarmi(data.StatusString);
            aggiornaModalita(data.OperativeModeString);
        })
        .catch(error => {
            console.error("Errore nell'aggiornare lo stato del sistema:", error);
            document.querySelectorAll('[data-device]').forEach(link => {
                const nome = link.getAttribute('data-device');
                const isOnline = false
                link.classList.toggle('online', isOnline);
                link.classList.toggle('offline', !isOnline);
            });
        })
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

document.addEventListener('DOMContentLoaded', () => {
    const NUM_PUNTI = 100;
    const initialLabels = Array.from({ length: NUM_PUNTI }, (_, i) => `T${i + 1}`);
    const initialData = {
        labels: initialLabels,
        datasets: [{
            label: 'Temperatura (°C)',
            data: Array.from({ length: NUM_PUNTI }, () => 0),
            borderColor: 'rgb(75, 192, 192)',
            tension: 0.1
        }]
    };

    const chartConfig = createChartConfig(initialData);
    const tempChart = new Chart(document.getElementById('tempChart'), chartConfig);

    update(tempChart);
    setInterval(() => update(tempChart), 100);

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