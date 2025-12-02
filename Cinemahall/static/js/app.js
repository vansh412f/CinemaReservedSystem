const seatMap = document.getElementById('seat-map');
const countSpan = document.getElementById('count');
const totalSpan = document.getElementById('total');
const payBtn = document.getElementById('pay-btn');
const timerContainer = document.getElementById('timer-container');
const timerSpan = document.getElementById('timer');
const toast = document.getElementById('toast');

// Views
const movieSelectionView = document.getElementById('movie-selection');
const seatSelectionView = document.getElementById('seat-selection');
const movieList = document.getElementById('movie-list');
const pageTitle = document.getElementById('page-title');
const pageSubtitle = document.getElementById('page-subtitle');
const backBtn = document.getElementById('back-btn');
const myTicketsBtn = document.getElementById('my-tickets-btn');

// Modal Elements
const bookingModal = document.getElementById('booking-modal');
const modalCloseBtn = document.getElementById('modal-close-btn');
const cardCode = document.getElementById('card-code');
const cardMovie = document.getElementById('card-movie');
const cardSeats = document.getElementById('card-seats');

let selectedSeats = []; // Array of seat objects {id, price}
let heldBookingID = null;
let timerInterval = null;
let currentShowID = null;
let currentMovieTitle = "";
let pollInterval = null;

// Initialize
fetchMovies();

backBtn.addEventListener('click', showMovieSelection);
myTicketsBtn.addEventListener('click', showMyTickets);

// Modal Close Logic
modalCloseBtn.addEventListener('click', (e) => {
    e.stopPropagation();
    closeModal();
});
bookingModal.addEventListener('click', (e) => {
    if (e.target === bookingModal) {
        closeModal();
    }
});

function closeModal() {
    console.log("Closing modal");
    bookingModal.classList.add('hidden');
    showMovieSelection();
}

async function fetchMovies() {
    try {
        const res = await fetch('/api/movies');
        const movies = await res.json();
        renderMovies(movies);
    } catch (err) {
        console.error('Error fetching movies:', err);
        showToast("Error loading movies");
    }
}

function renderMovies(movies) {
    movieList.innerHTML = '';
    movies.forEach(movie => {
        const card = document.createElement('div');
        card.className = 'movie-card';
        card.innerHTML = `
            <img src="${movie.poster_url}" alt="${movie.title}" class="movie-poster">
            <div class="movie-info">
                <div class="movie-title">${movie.title}</div>
                <div class="movie-duration">${Math.floor(movie.duration / 60)}h ${movie.duration % 60}m</div>
            </div>
        `;
        card.addEventListener('click', () => selectMovie(movie));
        movieList.appendChild(card);
    });
}

function selectMovie(movie) {
    currentShowID = movie.show_id;
    currentMovieTitle = movie.title;

    // Update Header
    pageTitle.innerText = movie.title;
    pageSubtitle.innerText = "IMAX Hall • 19:00";

    // Switch View
    movieSelectionView.classList.add('hidden');
    seatSelectionView.classList.remove('hidden');


    // Reset State
    selectedSeats = [];
    updateSummary();
    payBtn.disabled = false;
    payBtn.innerText = "Proceed to Pay";
    payBtn.onclick = handlePayClick; // Reset handler
    timerContainer.classList.add('hidden');
    clearInterval(timerInterval);

    // Load Seats
    fetchSeats();
    // Start Polling
    if (pollInterval) clearInterval(pollInterval);
    pollInterval = setInterval(fetchSeats, 5000);
}

function showMovieSelection() {
    pageTitle.innerText = "Cinema Hall";
    pageSubtitle.innerText = "Select a Movie";

    movieSelectionView.classList.remove('hidden');
    seatSelectionView.classList.add('hidden');

    if (pollInterval) clearInterval(pollInterval);
}

async function showMyTickets() {
    window.location.href = 'tickets.html';
}



async function fetchSeats() {
    if (!currentShowID) return;
    try {
        const res = await fetch(`/api/seats?show_id=${currentShowID}`);
        const seats = await res.json();
        renderSeats(seats);
    } catch (err) {
        console.error('Error fetching seats:', err);
    }
}

function renderSeats(seats) {
    const rows = {};
    seats.forEach(seat => {
        if (!rows[seat.row]) rows[seat.row] = [];
        rows[seat.row].push(seat);
    });

    if (seatMap.children.length === 0) {
        buildMap(rows);
    } else {
        updateMap(seats);
    }
}

function buildMap(rows) {
    seatMap.innerHTML = '';
    const sortedRows = Object.keys(rows).sort();
    let lastCategory = '';

    sortedRows.forEach(rowLabel => {
        const rowSeats = rows[rowLabel];
        const category = rowSeats[0].category;

        if (category !== lastCategory && lastCategory !== '') {
            const gap = document.createElement('div');
            gap.className = 'seat-gap';
            gap.textContent = `${category} - $${rowSeats[0].price}`;
            seatMap.appendChild(gap);
        }
        lastCategory = category;

        const rowDiv = document.createElement('div');
        rowDiv.className = 'row';

        const labelDiv = document.createElement('div');
        labelDiv.className = 'row-label';
        labelDiv.textContent = rowLabel;
        rowDiv.appendChild(labelDiv);

        rowSeats.forEach(seat => {
            const seatDiv = document.createElement('div');
            seatDiv.className = `seat ${seat.status.toLowerCase()}`;
            seatDiv.dataset.id = seat.id;

            seatDiv.addEventListener('click', () => handleSeatClick(seat, seatDiv));

            rowDiv.appendChild(seatDiv);
        });

        seatMap.appendChild(rowDiv);
    });
}

function updateMap(seats) {
    seats.forEach(seat => {
        const seatDiv = document.querySelector(`.seat[data-id="${seat.id}"]`);
        if (seatDiv) {
            const newStatus = seat.status.toLowerCase();
            const isSelectedByMe = selectedSeats.some(s => s.id === seat.id);

            if (isSelectedByMe && (newStatus === 'sold' || newStatus === 'held')) {
                if (!heldBookingID) {
                    deselectSeat(seat.id);
                    showToast(`Seat ${seat.row}${seat.number} just taken!`);
                }
            }

            if (isSelectedByMe && !heldBookingID) {
                seatDiv.className = `seat selected`;
            } else if (heldBookingID && isSelectedByMe) {
                seatDiv.className = `seat selected`;
            } else {
                seatDiv.className = `seat ${newStatus}`;
            }
        }
    });
}

function handleSeatClick(seat, el) {
    if (el.classList.contains('sold') || el.classList.contains('held')) return;

    const index = selectedSeats.findIndex(s => s.id === seat.id);
    if (index === -1) {
        selectedSeats.push({ id: seat.id, price: seat.price });
        el.classList.add('selected');
        el.classList.remove('available');
    } else {
        selectedSeats.splice(index, 1);
        el.classList.remove('selected');
        el.classList.add('available');
    }
    updateSummary();
}

function deselectSeat(id) {
    const index = selectedSeats.findIndex(s => s.id === id);
    if (index !== -1) {
        selectedSeats.splice(index, 1);
        updateSummary();
        const el = document.querySelector(`.seat[data-id="${id}"]`);
        if (el) el.classList.remove('selected');
    }
}

function updateSummary() {
    countSpan.innerText = selectedSeats.length;
    const total = selectedSeats.reduce((sum, s) => sum + s.price, 0);
    totalSpan.innerText = total;
}

async function handlePayClick() {
    if (selectedSeats.length === 0) {
        showToast("Please select seats first");
        return;
    }

    payBtn.disabled = true;
    payBtn.innerText = "Processing...";

    try {
        const res = await fetch('/api/hold-seats', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                show_id: currentShowID,
                seat_ids: selectedSeats.map(s => s.id),
                user_email: "test@example.com"
            })
        });

        if (res.status === 409) {
            showToast("One or more seats are no longer available");
            payBtn.disabled = false;
            payBtn.innerText = "Proceed to Pay";
            fetchSeats();
            return;
        }

        if (!res.ok) throw new Error('Failed to hold seats');

        const data = await res.json();
        heldBookingID = data.booking_id;

        startTimer(data.expires_at);
        showToast("Seats held! Complete payment in 5 minutes.");

        payBtn.innerText = "Confirm Payment";
        payBtn.disabled = false;
        payBtn.onclick = confirmBooking;

    } catch (err) {
        console.error(err);
        showToast("Error processing request");
        payBtn.disabled = false;
        payBtn.innerText = "Proceed to Pay";
    }
}

// Initial binding
payBtn.onclick = handlePayClick;

async function confirmBooking() {
    payBtn.disabled = true;
    payBtn.innerText = "Confirming...";

    try {
        const res = await fetch('/api/confirm-booking', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ booking_id: heldBookingID })
        });

        if (!res.ok) throw new Error('Failed to confirm');

        const data = await res.json();

        showBookingSuccess(data);

    } catch (err) {
        console.error(err);
        showToast("Error confirming booking");
        payBtn.disabled = false;
        payBtn.innerText = "Confirm Payment";
    }
}

function showBookingSuccess(data) {
    clearInterval(timerInterval);
    if (pollInterval) clearInterval(pollInterval);

    // Populate Card
    cardCode.innerText = data.booking_code;
    cardMovie.innerText = data.movie_title;
    cardSeats.innerText = data.seats.join(', ');

    // Show Modal
    bookingModal.classList.remove('hidden');

    showToast("Booking Confirmed!");
}

function startTimer(expiryStr) {
    timerContainer.classList.remove('hidden');
    const expiry = new Date(expiryStr).getTime();

    timerInterval = setInterval(() => {
        const now = new Date().getTime();
        const distance = expiry - now;

        if (distance < 0) {
            clearInterval(timerInterval);
            timerSpan.innerText = "EXPIRED";
            alert("Session Expired");
            location.reload();
            return;
        }

        const minutes = Math.floor((distance % (1000 * 60 * 60)) / (1000 * 60));
        const seconds = Math.floor((distance % (1000 * 60)) / 1000);

        timerSpan.innerText = `${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
    }, 1000);
}

function showToast(msg) {
    toast.textContent = msg;
    toast.classList.remove('hidden');
    setTimeout(() => {
        toast.classList.add('hidden');
    }, 3000);
}
