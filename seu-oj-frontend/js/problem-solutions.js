function solutionVisibilityClass(visibility) {
  if (visibility === "public") return "status-accepted";
  if (visibility === "class") return "status-pending";
  return "status-neutral";
}

function renderSolutionEmpty(message) {
  return `<div class="solution-empty">${escapeHTML(message)}</div>`;
}

function renderSolutionPreview(content) {
  if (typeof renderMarkdownBlock === "function") {
    return renderMarkdownBlock(content);
  }
  return `<div class="problem-markdown">${escapeHTML(content || "No content.")}</div>`;
}

function solutionDraftForUser(list) {
  return list.find((item) => Number(item.author_id) === Number(state.user?.id)) || null;
}

function renderSolutionForm(problemID, solution) {
  const isEditing = !!solution;
  return `
    <form id="problem-solution-form" class="solution-form" data-solution-id="${isEditing ? solution.id : ""}" data-author-id="${isEditing ? solution.author_id : ""}">
      <input type="hidden" name="solution_id" value="${isEditing ? solution.id : ""}" />
      <label class="field-label">Title</label>
      <input class="text-input" name="title" value="${escapeHTML(solution?.title || "")}" required />
      <label class="field-label">Visibility</label>
      <select class="select-input" name="visibility">
        <option value="public" ${solution?.visibility === "public" ? "selected" : ""}>public</option>
        <option value="private" ${solution?.visibility === "private" ? "selected" : ""}>private</option>
      </select>
      <label class="field-label">Content</label>
      <textarea class="text-area solution-editor-source" name="content" rows="20" required>${escapeHTML(solution?.content || "")}</textarea>
      <div class="view-subtitle">Markdown is supported. Non-admin users need an Accepted submission on this problem before publishing.</div>
      <div class="solution-form-actions">
        <button class="primary-button" type="submit">${isEditing ? "Save Solution" : "Create Solution"}</button>
        ${isEditing ? `<button class="ghost-button solution-delete-current" type="button" data-solution-id="${solution.id}">Delete</button>` : ""}
        <a class="ghost-button" href="#/problems/${problemID}">Cancel</a>
      </div>
    </form>
  `;
}

function renderAdminSolutionList(list) {
  if (!state.user || state.user.role !== "admin") return "";
  return `
    <div class="solution-admin-list">
      <div class="solution-admin-list-head">
        <h4>All Solutions</h4>
        <span class="status-pill status-neutral">${list.length}</span>
      </div>
      ${list.length ? `
        <div class="solution-admin-items">
          ${list.map((item) => `
            <article class="solution-admin-item">
              <div>
                <strong>${escapeHTML(item.title)}</strong>
                <p class="view-subtitle">Author #${item.author_id} / updated ${escapeHTML(item.updated_at)}</p>
              </div>
              <div class="solution-admin-item-actions">
                <span class="status-pill ${solutionVisibilityClass(item.visibility)}">${escapeHTML(item.visibility)}</span>
                <button class="ghost-button solution-load-button" type="button"
                  data-solution-id="${item.id}"
                  data-author-id="${item.author_id}"
                  data-title="${encodeURIComponent(item.title || "")}"
                  data-visibility="${escapeHTML(item.visibility || "public")}"
                  data-content="${encodeURIComponent(item.content || "")}">Edit</button>
                <button class="ghost-button solution-delete-button" type="button" data-solution-id="${item.id}">Delete</button>
              </div>
            </article>
          `).join("")}
        </div>
      ` : renderSolutionEmpty("No solutions yet.")}
    </div>
  `;
}

async function renderProblemSolutionManager(problemID, scope = "my") {
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
    const activeScope = canManageAll && scope === "all" ? "all" : "my";
    const ownSolution = solutionDraftForUser(list);
    const draftSolution = activeScope === "all" ? (ownSolution || list[0] || null) : (ownSolution || null);
    app.innerHTML = `
      <div class="solution-manager-page">
        <header class="solution-manager-bar">
          <div>
            <h1>My Solution</h1>
            <p>${escapeHTML(problem.title)} / #${problem.id}</p>
          </div>
          <div class="solution-manager-actions">
            ${canManageAll ? `
              <button class="ghost-button solution-scope-button ${activeScope === "my" ? "is-active" : ""}" type="button" data-solution-scope="my">My Solution</button>
              <button class="ghost-button solution-scope-button ${activeScope === "all" ? "is-active" : ""}" type="button" data-solution-scope="all">All Solutions</button>
            ` : ""}
            <a class="ghost-button" href="#/problems/${problem.id}">Back to Problem</a>
          </div>
        </header>
        <section class="solution-manager-layout">
          <section class="detail-card solution-manager-section solution-editor-section">
            <div class="solution-section-heading">
              <h3>${draftSolution ? "Edit" : "Create"}</h3>
              ${draftSolution ? '<span class="status-pill status-neutral">Existing solution</span>' : ""}
            </div>
            ${renderSolutionForm(problem.id, draftSolution)}
            ${activeScope === "all" ? renderAdminSolutionList(list) : ""}
          </section>
          <section class="detail-card solution-manager-section solution-preview-section">
            <div class="solution-section-heading">
              <h3>${activeScope === "all" ? "Selected Solution" : "My Solution"}</h3>
              <span class="view-subtitle">Preview</span>
            </div>
            <div class="solution-preview-title">${escapeHTML(draftSolution?.title || "Untitled Solution")}</div>
            <div class="solution-preview-meta">
              <span class="status-pill ${solutionVisibilityClass(draftSolution?.visibility || "public")}">${escapeHTML(draftSolution?.visibility || "public")}</span>
              ${draftSolution ? `<span>Author #${draftSolution.author_id}</span>` : "<span>Draft</span>"}
            </div>
            <div class="solution-preview-body">${renderSolutionPreview(draftSolution?.content || "")}</div>
          </section>
        </section>
      </div>
    `;

    setupSolutionEditor(problemID, activeScope);
  } catch (err) {
    app.innerHTML = `<div class="detail-card"><p>Load problem solutions failed: ${escapeHTML(err.message)}</p></div>`;
  }
}

function setupSolutionEditor(problemID, scope) {
  const form = document.getElementById("problem-solution-form");
  const titleInput = form?.querySelector('[name="title"]');
  const visibilityInput = form?.querySelector('[name="visibility"]');
  const contentInput = form?.querySelector('[name="content"]');
  const previewTitle = document.querySelector(".solution-preview-title");
  const previewMeta = document.querySelector(".solution-preview-meta");
  const previewBody = document.querySelector(".solution-preview-body");
  let markdownEditor = null;

  const refreshPreview = () => {
    const title = titleInput?.value.trim() || "Untitled Solution";
    const visibility = visibilityInput?.value || "public";
    const authorID = form?.dataset.authorId || "";
    if (previewTitle) previewTitle.textContent = title;
    if (previewMeta) {
      previewMeta.innerHTML = `
        <span class="status-pill ${solutionVisibilityClass(visibility)}">${escapeHTML(visibility)}</span>
        <span>${authorID ? `Author #${escapeHTML(authorID)}` : "Draft"}</span>
      `;
    }
    if (previewBody && contentInput) {
      previewBody.innerHTML = renderSolutionPreview(contentInput.value);
    }
  };

  [titleInput, visibilityInput, contentInput].forEach((input) => {
    input?.addEventListener("input", refreshPreview);
    input?.addEventListener("change", refreshPreview);
  });

  const attachMarkdownEditor = () => {
    if (!contentInput || typeof window.createSolutionMarkdownEditor !== "function") return;
    markdownEditor = window.createSolutionMarkdownEditor(contentInput, {
      onChange: refreshPreview,
    });
  };

  if (window.codeMirrorReadyPromise) {
    window.codeMirrorReadyPromise.then(attachMarkdownEditor).catch(() => {});
  } else {
    attachMarkdownEditor();
  }

  form?.addEventListener("submit", async (event) => {
    event.preventDefault();
    const formData = new FormData(form);
    const solutionID = String(form.dataset.solutionId || "").trim();
    const method = solutionID ? "PUT" : "POST";
    const path = solutionID ? `/problems/${problemID}/solutions/${solutionID}` : `/problems/${problemID}/solutions`;
    try {
      await apiFetch(path, {
        method,
        body: JSON.stringify({
          title: formData.get("title"),
          visibility: formData.get("visibility"),
          content: formData.get("content"),
        }),
      });
      setFlash(solutionID ? `Solution #${solutionID} updated` : "Solution created", false);
      return renderProblemSolutionManager(problemID, scope);
    } catch (err) {
      setFlash(err.message, true);
    }
  });

  document.querySelectorAll(".solution-load-button").forEach((button) => {
    button.addEventListener("click", () => {
      const solutionID = button.dataset.solutionId || "";
      form.dataset.solutionId = solutionID;
      form.dataset.authorId = button.dataset.authorId || "";
      form.querySelector('[name="solution_id"]').value = solutionID;
      titleInput.value = decodeURIComponent(button.dataset.title || "");
      visibilityInput.value = button.dataset.visibility || "public";
      const nextContent = decodeURIComponent(button.dataset.content || "");
      if (markdownEditor) {
        markdownEditor.setValue(nextContent);
      } else {
        contentInput.value = nextContent;
      }
      const submit = form.querySelector('button[type="submit"]');
      if (submit) submit.textContent = "Save Solution";
      refreshPreview();
    });
  });

  document.querySelectorAll(".solution-scope-button").forEach((button) => {
    button.addEventListener("click", () => {
      const nextScope = button.dataset.solutionScope || "my";
      renderProblemSolutionManager(problemID, nextScope);
    });
  });

  document.querySelectorAll(".solution-delete-button, .solution-delete-current").forEach((button) => {
    button.addEventListener("click", async () => {
      const solutionID = button.dataset.solutionId;
      const confirmed = window.confirm(`Delete solution #${solutionID}?`);
      if (!confirmed) return;
      try {
        await apiFetch(`/problems/${problemID}/solutions/${solutionID}`, { method: "DELETE" });
        setFlash(`Solution #${solutionID} deleted`, false);
        return renderProblemSolutionManager(problemID, scope);
      } catch (err) {
        setFlash(err.message, true);
      }
    });
  });
}
