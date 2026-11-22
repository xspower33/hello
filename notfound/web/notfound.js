(function () {
  var panel = document.getElementById("panel");
  var codeEl = panel ? panel.querySelector("code") : null;
  var chips = document.getElementById("chips");
  var roll = document.getElementById("roll");

  function show(text, isError) {
    if (!panel || !codeEl) return;
    codeEl.textContent = text;
    panel.classList.toggle("is-error", Boolean(isError));
    panel.hidden = false;
    // animasyonu yeniden tetikle
    panel.style.animation = "none";
    void panel.offsetWidth;
    panel.style.animation = "";
  }

  function call(path) {
    show("› " + path + "\n\nbekleniyor...");
    fetch(path, { headers: { Accept: "application/json" } })
      .then(function (res) {
        return res.text().then(function (body) {
          return { ok: res.ok, status: res.status, body: body };
        });
      })
      .then(function (r) {
        var pretty = r.body;
        try {
          pretty = JSON.stringify(JSON.parse(r.body), null, 2);
        } catch (e) {
          /* JSON değilse olduğu gibi bırak */
        }
        show("› " + path + "\n\n" + pretty, !r.ok);
      })
      .catch(function (err) {
        show("› " + path + "\n\nhata: " + err.message, true);
      });
  }

  if (chips) {
    chips.addEventListener("click", function (event) {
      var button = event.target.closest("button[data-path]");
      if (button) call(button.getAttribute("data-path"));
    });
  }

  if (roll) {
    roll.addEventListener("click", function () {
      show("› /random/endpoint\n\nzar atılıyor...");
      fetch("/random/endpoint")
        .then(function (res) {
          if (!res.ok) throw new Error("HTTP " + res.status);
          return res.json();
        })
        .then(function (data) {
          call(data.path);
        })
        .catch(function (err) {
          show("› /random/endpoint\n\nhata: " + err.message, true);
        });
    });
  }
})();
