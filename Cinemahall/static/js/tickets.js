const ticketsList = document.getElementById('tickets-list');
const backHomeBtn = document.getElementById('back-home-btn');
const toast = document.getElementById('toast');

backHomeBtn.addEventListener('click', () => {
    window.location.href = 'index.html';
});

// Initialize
fetchTickets();

async function fetchTickets() {
    try {
        const res = await fetch('/api/my-bookings');
        if (!res.ok) throw new Error('Failed to fetch bookings');
        const bookings = await res.json();
        renderTickets(bookings);
    } catch (err) {
        console.error(err);
        showToast("Error loading tickets");
        ticketsList.innerHTML = '<p>Error loading tickets. Please try again.</p>';
    }
}

function renderTickets(bookings) {
    ticketsList.innerHTML = '';
    if (!bookings || bookings.length === 0) {
        ticketsList.innerHTML = '<p>No bookings found.</p>';
        return;
    }

    bookings.forEach(b => {
        const card = document.createElement('div');
        card.className = 'ticket-card';
        card.innerHTML = `
            <h3>${b.movie_title}</h3>
            <p>Code: <strong>${b.booking_code}</strong></p>
            <p>Seats: <strong>${b.seats.join(', ')}</strong></p>
            <p>Date: <strong>${b.date}</strong></p>
            <p>IMAX Hall</p>
        `;
        ticketsList.appendChild(card);
    });
}

function showToast(msg) {
    toast.textContent = msg;
    toast.classList.remove('hidden');
    setTimeout(() => {
        toast.classList.add('hidden');
    }, 3000);
}
