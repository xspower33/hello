(function () {
  var months = [
    "Ocak", "Şubat", "Mart", "Nisan", "Mayıs", "Haziran",
    "Temmuz", "Ağustos", "Eylül", "Ekim", "Kasım", "Aralık",
  ];
  var days = [
    "Pazar", "Pazartesi", "Salı", "Çarşamba", "Perşembe", "Cuma", "Cumartesi",
  ];

  var clock = document.getElementById("clock");
  var dateEl = document.getElementById("date");
  var greetEl = document.getElementById("greeting");
  var srEl = document.querySelector(".sr-only");

  function greeting(hour) {
    if (hour >= 5 && hour < 12) return "günaydın";
    if (hour >= 12 && hour < 18) return "iyi günler";
    if (hour >= 18 && hour < 23) return "iyi akşamlar";
    return "iyi geceler";
  }

  function pad(n) {
    return n < 10 ? "0" + n : String(n);
  }

  function render() {
    var now = new Date();
    var time = pad(now.getHours()) + ":" + pad(now.getMinutes()) + ":" + pad(now.getSeconds());
    var date = now.getDate() + " " + months[now.getMonth()] + " " + now.getFullYear() +
      ", " + days[now.getDay()];

    if (clock) clock.textContent = time;
    if (srEl) srEl.textContent = "Şu anki saat: " + time;
    if (dateEl) dateEl.textContent = date;
    if (greetEl) greetEl.textContent = greeting(now.getHours());
  }

  render();
  setInterval(render, 1000);
})();
