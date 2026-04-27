function solutionVisibilityClass(visibility) {
  if (visibility === "public") return "status-accepted";
  if (visibility === "class") return "status-pending";
  return "status-neutral";
}

function renderSolutionEmpty(message) {
  return `<div class="solution-empty">${escapeHTML(message)}</div>`;
}

async function renderProblemSolutionManager(problemID) {
  if (!state.user) {
    app.innerHTML = `<div class="detail-card"><p>Login required.</p></div>`;
    return;
  }
  app.innerHTML = `<div class="detail-card"><p>Loading problem solutions...</p></div>`;
  try {
    const problemPath = state.user?.role === "admin" ? `/admin/problems/${problemID}` : `/problems/${problemID}`;
    const [problem, solutions] = await Promise.all([
      apiFetch(problemPath, { method: "GET" }),
      apiFetch(`/problems/${problemID}/solutions/manage`, { method: "GET" }),
    ]);
    const list = solutions.list || [];
    const canManageAll = state.user?.role === "admin";
    const ownSolution = list.find((item) => Number(item.author_id) === Number(state.user?.id));
    const canCreate = !ownSolution;
    app.innerHTML = `
      <div class="view-header solution-manager-header">
        <div>
          <h1 class="view-title">My Solutions</h1>
          <p class="view-subtitle">${escapeHTML(problem.title)} / #${problem.id}</p>
        </div>
        <div class="solution-manager-actions">
          <a class="ghost-button" href="#/problems/${problem.id}">Back to Problem</a>
          ${state.user?.role === "admin" ? `<a class="ghost-button" href="#/admin/problems/${problem.id}">Admin View</a>` : ""}
        </div>
      </div>
      <section class="solution-manager-layout">
        <section class="detail-card solution-manager-section">
          <div class="solution-section-heading">
            <h3>Create</h3>
            ${ownSolution ? '<span class="status-pill status-neutral">Already created</span>' : ""}
          </div>
          ${canCreate ? `
            <form id="problem-solution-form" class="solution-form">
              <label class="field-label">Title</label>
              <input class="text-input" name="title" required />
              <label class="field-label">Visibility</label>
              <select class="select-input" name="visibility">
                <option value="public">public</option>
                <option value="private">private</option>
              </select>
              <label class="field-label">Content</label>
              <div class="solution-markdown-editor">
                <div class="solution-preview-tabs" role="tablist" aria-label="Create solution content mode">
                  <button class="solution-preview-tab is-active" type="button" data-preview-mode="edit" data-preview-group="create">Edit</button>
                  <button class="solution-preview-tab" type="button" data-preview-mode="preview" data-preview-group="create">Preview</button>
                </div>
                <textarea class="text-area solution-markdown-source" name="content" rows="18" required data-preview-source="create"></textarea>
                <div class="solution-markdown-preview hidden" data-preview-target="create">${renderSolutionPreview("")}</div>
              </div>
              <div class="view-subtitle">Markdown is supported. Non-admin users need an Accepted submission on this problem before publishing.</div>
              <div class="solution-form-actions"><button class="primary-button" type="submit">Create Solution</button></div>
            </form>
          ` : `
            <div class="solution-empty">
              You already have a solution for this problem. Edit it in the section on the right.
            </div>
          `}
        </section>
        <section class="detail-card solution-manager-section">
          <div class="solution-section-heading">
            <h3>Your Solution</h3>
            ${canManageAll ? '<span class="view-subtitle">Admin view shows all authors.</span>' : ""}
          </div>
          ${list.length ? `
            <div class="solution-stack">
              ${list.map((item) => `
                <article class="solution-card">
                  <div class="view-header compact">
                    <div>
                      <h4 class="solution-title">${escapeHTML(item.title)}</h4>
                      <p class="view-subtitle">Author #${item.author_id} / updated ${escapeHTML(item.updated_at)}</p>
                    </div>
                    <span class="status-pill ${solutionVisibilityClass(item.visibility)}">${escapeHTML(item.visibility)}</span>
                  </div>
                  <form class="solution-edit-form" data-solution-id="${item.id}">
                    <label class="field-label">Title</label>
                    <input class="text-input" name="title" value="${escapeHTML(item.title)}" required />
                    <label class="field-label">Visibility</label>
                    <select class="select-input" name="visibility">
                      <option value="public" ${item.visibility === "public" ? "selected" : ""}>public</option>
                      <option value="private" ${item.visibility === "private" ? "selected" : ""}>private</option>
                    </select>
                    <label class="field-label">Content</label>
                    <textarea class="text-area" name="content" rows="10" required>${escapeHTML(item.content || "")}</textarea>
                    <div class="solution-form-actions">
                      <button class="ghost-button" type="submit">Save</button>
                      <button class="ghost-button solution-delete-button" type="button" data-solution-id="${item.id}">Delete</button>
                    </div>
                  </form>
                </article>
              `).join("")}
            </div>
          ` : renderSolutionEmpty("No solutions yet.")}
        </section>
      </section>
    `;

    setupSolutionMarkdownPreview();

    document.getElementById("problem-solution-form")?.addEventListener("submit", async (event) => {
      event.preventDefault();
      const form = new FormData(event.currentTarget);
      try {
        await apiFetch(`/problems/${problemID}/solutions`, {
          method: "POST",
          body: JSON.stringify({
            title: form.get("title"),
            visibility: form.get("visibility"),
            content: form.get("content"),
          }),
        });
        setFlash("Solution created", false);
        return renderProblemSolutionManager(problemID);
      } catch (err) {
        setFlash(err.message, true);
      }
    });

    document.querySelectorAll(".solution-edit-form").forEach((formElement) => {
      formElement.addEventListener("submit", async (event) => {
        event.preventDefault();
        const form = new FormData(formElement);
        const solutionID = formElement.dataset.solutionId;
        try {
          await apiFetch(`/problems/${problemID}/solutions/${solutionID}`, {
            method: "PUT",
            body: JSON.stringify({
              title: form.get("title"),
              visibility: form.get("visibility"),
              content: form.get("content"),
            }),
          });
          setFlash(`Solution #${solutionID} updated`, false);
          return renderProblemSolutionManager(problemID);
        } catch (err) {
          setFlash(err.message, true);
        }
      });
    });

    document.querySelectorAll(".solution-delete-button").forEach((button) => {
      button.addEventListener("click", async () => {
        const solutionID = button.dataset.solutionId;
        const confirmed = window.confirm(`Delete solution #${solutionID}?`);
        if (!confirmed) return;
        try {
          await apiFetch(`/problems/${problemID}/solutions/${solutionID}`, { method: "DELETE" });
          setFlash(`Solution #${solutionID} deleted`, false);
          return renderProblemSolutionManager(problemID);
        } catch (err) {
          setFlash(err.message, true);
        }
      });
    });
  } catch (err) {
    app.innerHTML = `<div class="detail-card"><p>Load problem solutions failed: ${escapeHTML(err.message)}</p></div>`;
  }
}

function renderSolutionPreview(content) {
  if (typeof renderMarkdownBlock === "function") {
    return renderMarkdownBlock(content);
  }
  return `<div class="problem-markdown">${escapeHTML(content || "No content.")}</div>`;
}

function setupSolutionMarkdownPreview() {
  document.querySelectorAll(".solution-preview-tab").forEach((button) => {
    button.addEventListener("click", () => {
      const group = button.dataset.previewGroup;
      const mode = button.dataset.previewMode;
      const source = document.querySelector(`[data-preview-source="${group}"]`);
      const target = document.querySelector(`[data-preview-target="${group}"]`);
      if (!source || !target) return;

      document.querySelectorAll(`[data-preview-group="${group}"]`).forEach((tab) => {
        tab.classList.toggle("is-active", tab === button);
      });

      if (mode === "preview") {
        target.innerHTML = renderSolutionPreview(source.value);
        source.classList.add("hidden");
        target.classList.remove("hidden");
      } else {
        target.classList.add("hidden");
        source.classList.remove("hidden");
        source.focus();
      }
    });
  });

  document.querySelectorAll(".solution-markdown-source").forEach((source) => {
    source.addEventListener("input", () => {
      const group = source.dataset.previewSource;
      const target = document.querySelector(`[data-preview-target="${group}"]`);
      if (target && !target.classList.contains("hidden")) {
        target.innerHTML = renderSolutionPreview(source.value);
      }
    });
  });
}
