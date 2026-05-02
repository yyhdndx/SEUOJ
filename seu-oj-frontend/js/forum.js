function forumScopeLabel(scopeType, scopeID) {
  if (scopeType === "problem") return scopeID ? `Problem #${scopeID}` : "Problem";
  if (scopeType === "contest") return scopeID ? `Contest #${scopeID}` : "Contest";
  return "General";
}

function forumScopeLink(scopeType, scopeID) {
  if (scopeType === "general") return "#/forum?scope_type=general";
  if (scopeType === "problem" && !scopeID) return "#/forum?scope_type=problem";
  if (scopeType === "contest" && !scopeID) return "#/forum?scope_type=contest";
  if (!scopeID) return "#/forum";
  if (scopeType === "problem") return `#/problems/${scopeID}`;
  if (scopeType === "contest") return `#/contests/${scopeID}`;
  return "#/forum";
}

function canManageForumContent(authorID) {
  return !!state.user && (Number(state.user.id) === Number(authorID) || state.user.role === "admin" || state.user.role === "teacher");
}

function canModerateForum() {
  return !!state.user && (state.user.role === "admin" || state.user.role === "teacher");
}

function forumStatusPills(topic) {
  const pills = [];
  const scopeLabel = forumScopeLabel(topic.scope_type, topic.scope_id);
  const scopeHref = forumScopeLink(topic.scope_type, topic.scope_id);
  pills.push(`<a class="status-pill status-neutral forum-scope-pill" href="${scopeHref}">${escapeHTML(scopeLabel)}</a>`);
  if (topic.is_pinned) pills.push(`<span class="status-pill status-accepted">Pinned</span>`);
  if (topic.is_locked) pills.push(`<span class="status-pill status-pending">Locked</span>`);
  return pills.join("");
}

function formatForumDate(iso) {
  if (!iso) return "-";
  const d = new Date(iso);
  if (isNaN(d.getTime())) return "-";
  const yyyy = d.getFullYear();
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  return `${yyyy}-${mm}-${dd}`;
}

function forumMetaLine(topic) {
  const created = formatForumDate(topic.created_at);
  return `
    <div class="forum-topic-meta">
      <span class="forum-meta-item"><strong>${escapeHTML(topic.author_name || "-")}</strong></span>
      <span class="forum-meta-sep">·</span>
      <span class="forum-meta-item">${created}</span>
    </div>
  `;
}

function forumStatsPills(topic) {
  const replies = Number(topic.reply_count) || 0;
  return `<span class="status-pill status-neutral">${replies} repl${replies === 1 ? "y" : "ies"}</span>`;
}

function renderForumSummaryCards(result, scopeType, scopeID) {
  const list = result?.list || [];
  const shown = list.length;
  const pinned = list.filter((item) => item.is_pinned).length;
  const locked = list.filter((item) => item.is_locked).length;
  const active = list.filter((item) => item.reply_count > 0).length;
  const scopeLabel = scopeType ? forumScopeLabel(scopeType, scopeID || null) : "All Topics";
  return `
    <div class="forum-summary-grid">
      <article class="forum-summary-card">
        <span class="forum-summary-label">Zone</span>
        <strong class="forum-summary-value">${escapeHTML(scopeLabel)}</strong>
      </article>
      <article class="forum-summary-card">
        <span class="forum-summary-label">Topics</span>
        <strong class="forum-summary-value">${shown}</strong>
      </article>
      <article class="forum-summary-card">
        <span class="forum-summary-label">Pinned</span>
        <strong class="forum-summary-value">${pinned}</strong>
      </article>
      <article class="forum-summary-card">
        <span class="forum-summary-label">Active</span>
        <strong class="forum-summary-value">${active}</strong>
      </article>
      <article class="forum-summary-card">
        <span class="forum-summary-label">Locked</span>
        <strong class="forum-summary-value">${locked}</strong>
      </article>
    </div>
  `;
}

function renderMarkdownPreview(source, maxChars) {
  const text = (source || "").toString().trim();
  if (!text) return "";
  const limit = maxChars || 400;
  if (text.length <= limit) return renderMarkdown(text);
  let truncated = text.substring(0, limit);
  const lastNewline = truncated.lastIndexOf("\n");
  if (lastNewline > limit * 0.6) {
    truncated = truncated.substring(0, lastNewline);
  } else {
    const lastSpace = truncated.lastIndexOf(" ");
    if (lastSpace > limit * 0.8) truncated = truncated.substring(0, lastSpace);
  }
  return renderMarkdown(truncated.trimEnd()) + `<span class="forum-preview-ellipsis">&hellip;</span>`;
}

function renderForumTopicCards(list) {
  if (!list.length) {
    return `<div class="detail-card"><p>No topics yet.</p></div>`;
  }
  return `
    <div class="forum-topic-list">
      ${list.map((item) => `
        <article class="forum-topic-card ${item.is_pinned ? "pinned" : ""} ${item.is_locked ? "locked" : ""}" data-topic-id="${item.id}">
          <h3 class="forum-topic-title">${escapeHTML(item.title)}</h3>
          ${forumMetaLine(item)}
          <div class="pill-row forum-topic-pill-row">
            ${forumStatusPills(item)}
            ${forumStatsPills(item)}
          </div>
          ${item.content_preview ? `<div class="forum-topic-preview">${renderMarkdownPreview(item.content_preview)}</div>` : ""}
          <div class="forum-topic-actions">
            <button class="forum-like-btn ${item.is_liked ? 'is-active' : ''}" data-topic-id="${item.id}" data-action="like" aria-label="Like">&#9825; <span class="forum-like-count">${item.like_count || 0}</span></button>
            <button class="forum-fav-btn ${item.is_favorited ? 'is-active' : ''}" data-topic-id="${item.id}" data-action="favorite" aria-label="Favorite">&#9733; <span class="forum-fav-count">${item.favorite_count || 0}</span></button>
          </div>
        </article>
      `).join("")}
    </div>
  `;
}

function openForumComposeModal(scopeType, scopeID) {
  const overlay = document.createElement("div");
  overlay.id = "forum-compose-overlay";
  overlay.className = "forum-compose-overlay";
  overlay.innerHTML = `
    <div class="forum-compose-modal">
      <div class="forum-compose-modal-header">
        <h3 class="view-title" style="margin:0;">New Topic</h3>
        <button class="forum-compose-modal-close ghost-button" type="button" aria-label="Close">&times;</button>
      </div>
      <form id="forum-compose-modal-form">
        <div id="forum-compose-modal-error" class="flash error" style="display:none; margin-bottom:14px;"></div>
        <label class="field-label">Title</label>
        <input class="text-input" name="title" required />
        <label class="field-label">Zone</label>
        <select class="select-input" name="scope_type">
          <option value="general" ${scopeType === "general" || !scopeType ? "selected" : ""}>general</option>
          <option value="problem" ${scopeType === "problem" ? "selected" : ""}>problem</option>
          <option value="contest" ${scopeType === "contest" ? "selected" : ""}>contest</option>
        </select>
        <label class="field-label">Content</label>
        <textarea class="text-area" name="content" rows="10" required></textarea>
        <div class="view-subtitle" style="margin-top:8px;">Markdown is supported. Use fenced code blocks for snippets and keep clarifications concise.</div>
        <div style="margin-top:14px; display:flex; gap:10px;">
          <button class="primary-button" type="submit">Post Topic</button>
          <button class="ghost-button" type="button" id="forum-compose-modal-cancel">Cancel</button>
        </div>
      </form>
    </div>
  `;

  function dismiss() {
    closeForumComposeModal(overlay, handleKey);
  }

  function handleKey(e) {
    if (e.key === "Escape") dismiss();
  }

  overlay.addEventListener("click", (e) => {
    if (e.target === overlay) dismiss();
  });
  overlay.querySelector(".forum-compose-modal-close")?.addEventListener("click", dismiss);
  overlay.querySelector("#forum-compose-modal-cancel")?.addEventListener("click", dismiss);
  document.addEventListener("keydown", handleKey);

  overlay.querySelector("#forum-compose-modal-form")?.addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const createScopeType = (form.get("scope_type") || "general").toString();
    const rawScopeID = (createScopeType === scopeType && scopeID) ? scopeID : "";
    try {
      const topic = await apiFetch("/forum/topics", {
        method: "POST",
        body: JSON.stringify({
          title: form.get("title"),
          content: form.get("content"),
          scope_type: createScopeType,
          scope_id: createScopeType === "general" || !rawScopeID ? null : Number(rawScopeID),
        }),
      });
      closeForumComposeModal(overlay, handleKey);
      setFlash(`Topic #${topic.id} created`, false);
      location.hash = `#/forum/topics/${topic.id}`;
    } catch (err) {
      const errorBox = overlay.querySelector("#forum-compose-modal-error");
      if (errorBox) {
        errorBox.textContent = err.message;
        errorBox.style.display = "block";
      }
    }
  });

  document.body.appendChild(overlay);
}

function closeForumComposeModal(overlay, handleKey) {
  const form = overlay?.querySelector("#forum-compose-modal-form");
  const title = form?.querySelector('[name="title"]')?.value?.trim();
  const content = form?.querySelector('[name="content"]')?.value?.trim();
  if (title || content) {
    if (!window.confirm("Discard unsaved content?")) return;
  }
  document.removeEventListener("keydown", handleKey);
  overlay?.remove();
}

async function toggleForumReaction(topicID, action, activate, btn) {
  const method = activate ? "POST" : "DELETE";
  try {
    await apiFetch(`/forum/topics/${topicID}/${action}`, { method });
    const countSpan = btn.querySelector("span");
    const current = Number(countSpan.textContent) || 0;
    if (activate) {
      btn.classList.add("is-active");
      countSpan.textContent = current + 1;
    } else {
      btn.classList.remove("is-active");
      countSpan.textContent = Math.max(current - 1, 0);
    }
  } catch (err) {
    setFlash(err.message, true);
  }
}

async function renderForum() {
  const query = getHashQueryParams();
  const page = Number(query.get("page") || 1);
  const pageSize = Number(query.get("page_size") || 20);
  const keyword = (query.get("keyword") || "").trim();
  const scopeType = (query.get("scope_type") || "").trim();
  const scopeID = (query.get("scope_id") || "").trim();

  app.innerHTML = `<div class="detail-card"><p>Loading forum...</p></div>`;
  try {
    const params = new URLSearchParams({ page: String(page), page_size: String(pageSize) });
    if (keyword) params.set("keyword", keyword);
    if (scopeType) params.set("scope_type", scopeType);
    if (scopeID) params.set("scope_id", scopeID);
    const topics = await apiFetch(`/forum/topics?${params.toString()}`, { method: "GET" });
    const topicList = topics.list || [];
    const totalTopics = Math.max(Number(topics.total ?? 0) || 0, topicList.length);
    const topicSummary = totalTopics > topicList.length
      ? `Showing ${topicList.length} of ${totalTopics} topics on this page.`
      : `Showing ${topicList.length} topic${topicList.length === 1 ? "" : "s"}.`;

    const scopeHint = scopeType
      ? `Current zone: ${forumScopeLabel(scopeType, scopeID || null)}. ${topicSummary}`
      : `General discussion across contests, problems, and training topics. ${topicSummary}`;

    const tabHref = (nextScopeType) => {
      const next = new URLSearchParams();
      if (keyword) next.set("keyword", keyword);
      if (nextScopeType) next.set("scope_type", nextScopeType);
      if ((nextScopeType === scopeType) && scopeID) {
        next.set("scope_id", scopeID);
      }
      next.set("page", "1");
      next.set("page_size", String(pageSize));
      const qs = next.toString();
      return qs ? `#/forum?${qs}` : "#/forum";
    };

    app.innerHTML = `
      <div class="view-header">
        <div>
          <h1 class="view-title">Forum</h1>
          <p class="view-subtitle">${escapeHTML(scopeHint)}</p>
        </div>
      </div>
      ${renderForumSummaryCards(topics, scopeType, scopeID)}
      <section style="margin-top:18px;">
        <article class="detail-card forum-panel-card">
          <div class="view-header compact">
            <div>
              <h3>Zones & Search</h3>
              <p class="view-subtitle">Filter topics by zone (general / problem / contest) and keyword.</p>
            </div>
            <div class="pill-row forum-scope-tabs" aria-label="Forum zones">
              <a class="ghost-button ${!scopeType ? "active-filter" : ""}" href="${tabHref("")}">All</a>
              <a class="ghost-button ${scopeType === "general" ? "active-filter" : ""}" href="${tabHref("general")}">General</a>
              <a class="ghost-button ${scopeType === "problem" ? "active-filter" : ""}" href="${tabHref("problem")}">Problem</a>
              <a class="ghost-button ${scopeType === "contest" ? "active-filter" : ""}" href="${tabHref("contest")}">Contest</a>
            </div>
          </div>
          <form id="forum-filter-form" class="forum-filter-form">
            <div class="forum-filter-grid">
              <div class="forum-filter-field keyword">
                <label class="field-label">Keyword</label>
                <input class="text-input" name="keyword" value="${escapeHTML(keyword)}" placeholder="Search title or content" />
              </div>
              <div class="forum-filter-field type">
                <label class="field-label">Zone</label>
                <select class="select-input" name="scope_type">
                  <option value="" ${!scopeType ? "selected" : ""}>all</option>
                  <option value="general" ${scopeType === "general" ? "selected" : ""}>general</option>
                  <option value="problem" ${scopeType === "problem" ? "selected" : ""}>problem</option>
                  <option value="contest" ${scopeType === "contest" ? "selected" : ""}>contest</option>
                </select>
              </div>
              <div class="forum-filter-field actions">
                <div class="forum-filter-actions">
                  <button class="primary-button" type="submit">Search</button>
                  <a class="ghost-button" href="#/forum">Reset</a>
                </div>
              </div>
            </div>
            <p class="view-subtitle" style="margin:10px 0 0;">Tip: open a problem or contest and click <strong>Discuss</strong> to jump into its zone automatically.</p>
          </form>
        </article>
      </section>
      <section style="margin-top:18px;">${renderForumTopicCards(topicList)}</section>
      ${state.user ? `<button id="forum-compose-fab" class="forum-compose-fab" title="New Topic"><span class="forum-compose-fab-icon">+</span></button>` : ""}
    `;

    document.getElementById("forum-filter-form")?.addEventListener("submit", (event) => {
      event.preventDefault();
      const form = new FormData(event.currentTarget);
      const next = new URLSearchParams();
      const nextKeyword = (form.get("keyword") || "").toString().trim();
      const nextScopeType = (form.get("scope_type") || "").toString().trim();
      if (nextKeyword) next.set("keyword", nextKeyword);
      if (nextScopeType) next.set("scope_type", nextScopeType);
      if (nextScopeType === scopeType && scopeID) next.set("scope_id", scopeID);
      next.set("page", "1");
      next.set("page_size", String(pageSize));
      location.hash = `#/forum?${next.toString()}`;
    });

    document.querySelector("#forum-filter-form input[name=keyword]")?.addEventListener("keydown", (event) => {
      if (event.key !== "Enter") return;
      event.preventDefault();
      document.getElementById("forum-filter-form")?.dispatchEvent(new Event("submit", { cancelable: true }));
    });

    document.getElementById("forum-compose-fab")?.addEventListener("click", () => {
      openForumComposeModal(scopeType, scopeID);
    });

    document.querySelector(".forum-topic-list")?.addEventListener("click", (event) => {
      const btn = event.target.closest(".forum-like-btn, .forum-fav-btn");
      if (btn) {
        event.stopPropagation();
        const topicID = btn.dataset.topicId;
        const action = btn.dataset.action;
        const isActive = btn.classList.contains("is-active");
        toggleForumReaction(topicID, action, !isActive, btn);
        return;
      }
      const card = event.target.closest(".forum-topic-card");
      if (!card) return;
      if (event.target.closest("a")) return;
      const topicID = card.dataset.topicId;
      if (topicID) location.hash = `#/forum/topics/${topicID}`;
    });
  } catch (err) {
    app.innerHTML = `<div class="detail-card"><p>Load forum failed: ${escapeHTML(err.message)}</p></div>`;
  }
}

async function renderForumTopicDetail(id) {
  app.innerHTML = `<div class="detail-card"><p>Loading topic...</p></div>`;
  try {
    const detail = await apiFetch(`/forum/topics/${id}`, { method: "GET" });
    const canManageTopic = canManageForumContent(detail.author_id);
    const canModerate = canModerateForum();

    app.innerHTML = `
      <div class="view-header">
        <div>
          <h1 class="view-title">${escapeHTML(detail.title)}</h1>
          <p class="view-subtitle">${escapeHTML(detail.author_name)} · ${escapeHTML(detail.created_at)} · ${escapeHTML(forumScopeLabel(detail.scope_type, detail.scope_id))}</p>
        </div>
        <div style="display:flex; gap:10px; flex-wrap:wrap;">
          <a class="ghost-button" href="#/forum?scope_type=${encodeURIComponent(detail.scope_type)}${detail.scope_id ? `&scope_id=${detail.scope_id}` : ""}">Back to Topics</a>
          <a class="ghost-button" href="${forumScopeLink(detail.scope_type, detail.scope_id)}">Open Scope</a>
          ${state.user ? `<button class="ghost-button forum-like-btn ${detail.is_liked ? 'is-active' : ''}" id="forum-detail-like-btn">&#9825; ${detail.is_liked ? 'Liked' : 'Like'} (<span id="forum-detail-like-count">${detail.like_count || 0}</span>)</button><button class="ghost-button forum-fav-btn ${detail.is_favorited ? 'is-active' : ''}" id="forum-detail-fav-btn">&#9733; ${detail.is_favorited ? 'Saved' : 'Save'} (<span id="forum-detail-fav-count">${detail.favorite_count || 0}</span>)</button>` : ""}
          ${canManageTopic ? `<button class="ghost-button" type="button" id="forum-topic-edit-btn">Edit Topic</button><button class="ghost-button" type="button" id="forum-topic-delete-btn">Delete Topic</button>` : ""}
        </div>
      </div>
      <section class="detail-grid" style="align-items:start;">
        <article class="detail-card forum-topic-detail-main">
          <div class="forum-topic-state-bar">
            ${forumStatusPills(detail)}
          </div>
          ${renderMarkdownBlock(detail.content || "")}
        </article>
        <aside class="detail-card forum-topic-sidebar">
          <h3>Topic Meta</h3>
          <div class="teaching-meta-grid">
            <div class="teaching-meta-card"><span class="teaching-meta-label">Scope</span><span class="teaching-meta-value">${escapeHTML(forumScopeLabel(detail.scope_type, detail.scope_id))}</span></div>
            <div class="teaching-meta-card"><span class="teaching-meta-label">Replies</span><span class="teaching-meta-value">${detail.reply_count}</span></div>
            <div class="teaching-meta-card"><span class="teaching-meta-label">Likes</span><span class="teaching-meta-value">&#9825; ${detail.like_count || 0}</span></div>
            <div class="teaching-meta-card"><span class="teaching-meta-label">Favorites</span><span class="teaching-meta-value">&#9733; ${detail.favorite_count || 0}</span></div>
            <div class="teaching-meta-card"><span class="teaching-meta-label">Created</span><span class="teaching-meta-value mono">${escapeHTML(detail.created_at)}</span></div>
            <div class="teaching-meta-card"><span class="teaching-meta-label">Updated</span><span class="teaching-meta-value mono">${escapeHTML(detail.updated_at)}</span></div>
          </div>
          ${canModerate ? `
            <div class="forum-moderation-box">
              <h4 style="margin:0 0 10px;">Moderation</h4>
              <div class="forum-topic-footer">
                <button class="ghost-button" type="button" id="forum-toggle-pin-btn">${detail.is_pinned ? "Unpin" : "Pin"}</button>
                <button class="ghost-button" type="button" id="forum-toggle-lock-btn">${detail.is_locked ? "Unlock" : "Lock"}</button>
              </div>
            </div>
          ` : ""}
        </aside>
      </section>
      <section class="detail-card" style="margin-top:18px;">
        <div class="view-header compact">
          <div>
            <h3>Replies</h3>
            <p class="view-subtitle">${detail.replies?.length || 0} reply(s)</p>
          </div>
          ${detail.is_locked ? `<span class="status-pill status-pending">Locked</span>` : ""}
        </div>
        ${(detail.replies || []).length ? `
          <div class="forum-reply-list">
            ${(detail.replies || []).map((item) => `
              <article class="forum-reply-card">
                <div class="view-header compact">
                  <div>
                    <strong>${escapeHTML(item.author_name)}</strong>
                    <p class="view-subtitle">${escapeHTML(item.created_at)}</p>
                  </div>
                  ${canManageForumContent(item.author_id) ? `<div style="display:flex; gap:8px; flex-wrap:wrap;"><button class="ghost-button forum-reply-edit" type="button" data-reply-id="${item.id}" data-reply-content="${encodeURIComponent(item.content || "")}">Edit</button><button class="ghost-button forum-reply-delete" type="button" data-reply-id="${item.id}">Delete</button></div>` : ""}
                </div>
                ${renderMarkdownBlock(item.content || "")}
              </article>
            `).join("")}
          </div>
        ` : `<p class="view-subtitle">No replies yet.</p>`}
      </section>
      <section class="detail-card" style="margin-top:18px;">
        <h3>Reply</h3>
        ${state.user ? `
          ${detail.is_locked && !canModerate ? `<p class="view-subtitle">This topic is locked. Only moderators can reply.</p>` : `
            <form id="forum-reply-form">
              <label class="field-label">Content</label>
              <textarea class="text-area" name="content" rows="8" required></textarea>
              <div class="view-subtitle" style="margin-top:8px;">Markdown is supported in replies.</div>
              <div style="margin-top:14px;"><button class="primary-button" type="submit">Post Reply</button></div>
            </form>
          `}
        ` : `<p class="view-subtitle">Login to reply.</p>`}
      </section>
    `;

    document.getElementById("forum-detail-like-btn")?.addEventListener("click", async () => {
      const btn = document.getElementById("forum-detail-like-btn");
      if (!btn) return;
      const activate = !btn.classList.contains("is-active");
      await toggleForumReaction(id, "like", activate, btn);
      return renderForumTopicDetail(id);
    });

    document.getElementById("forum-detail-fav-btn")?.addEventListener("click", async () => {
      const btn = document.getElementById("forum-detail-fav-btn");
      if (!btn) return;
      const activate = !btn.classList.contains("is-active");
      await toggleForumReaction(id, "favorite", activate, btn);
      return renderForumTopicDetail(id);
    });

    document.getElementById("forum-reply-form")?.addEventListener("submit", async (event) => {
      event.preventDefault();
      const form = new FormData(event.currentTarget);
      try {
        await apiFetch(`/forum/topics/${id}/replies`, {
          method: "POST",
          body: JSON.stringify({ content: form.get("content") }),
        });
        setFlash("Reply posted", false);
        return renderForumTopicDetail(id);
      } catch (err) {
        setFlash(err.message, true);
      }
    });

    document.getElementById("forum-topic-edit-btn")?.addEventListener("click", async () => {
      const nextTitle = window.prompt("Edit topic title", detail.title);
      if (nextTitle === null) return;
      const nextContent = window.prompt("Edit topic content", detail.content || "");
      if (nextContent === null) return;
      try {
        await apiFetch(`/forum/topics/${id}`, {
          method: "PUT",
          body: JSON.stringify({ title: nextTitle, content: nextContent }),
        });
        setFlash("Topic updated", false);
        return renderForumTopicDetail(id);
      } catch (err) {
        setFlash(err.message, true);
      }
    });

    document.getElementById("forum-topic-delete-btn")?.addEventListener("click", async () => {
      if (!window.confirm(`Delete topic #${id}?`)) return;
      try {
        await apiFetch(`/forum/topics/${id}`, { method: "DELETE" });
        setFlash("Topic deleted", false);
        location.hash = "#/forum";
      } catch (err) {
        setFlash(err.message, true);
      }
    });

    document.getElementById("forum-toggle-pin-btn")?.addEventListener("click", async () => {
      try {
        await apiFetch(`/forum/topics/${id}`, {
          method: "PUT",
          body: JSON.stringify({ title: detail.title, content: detail.content, is_pinned: !detail.is_pinned, is_locked: detail.is_locked }),
        });
        setFlash(detail.is_pinned ? "Topic unpinned" : "Topic pinned", false);
        return renderForumTopicDetail(id);
      } catch (err) {
        setFlash(err.message, true);
      }
    });

    document.getElementById("forum-toggle-lock-btn")?.addEventListener("click", async () => {
      try {
        await apiFetch(`/forum/topics/${id}`, {
          method: "PUT",
          body: JSON.stringify({ title: detail.title, content: detail.content, is_pinned: detail.is_pinned, is_locked: !detail.is_locked }),
        });
        setFlash(detail.is_locked ? "Topic unlocked" : "Topic locked", false);
        return renderForumTopicDetail(id);
      } catch (err) {
        setFlash(err.message, true);
      }
    });

    document.querySelectorAll(".forum-reply-edit").forEach((button) => {
      button.addEventListener("click", async () => {
        const replyID = button.dataset.replyId;
        const current = decodeURIComponent(button.dataset.replyContent || "");
        const nextContent = window.prompt("Edit reply", current);
        if (nextContent === null) return;
        try {
          await apiFetch(`/forum/replies/${replyID}`, {
            method: "PUT",
            body: JSON.stringify({ content: nextContent }),
          });
          setFlash(`Reply #${replyID} updated`, false);
          return renderForumTopicDetail(id);
        } catch (err) {
          setFlash(err.message, true);
        }
      });
    });

    document.querySelectorAll(".forum-reply-delete").forEach((button) => {
      button.addEventListener("click", async () => {
        const replyID = button.dataset.replyId;
        if (!window.confirm(`Delete reply #${replyID}?`)) return;
        try {
          await apiFetch(`/forum/replies/${replyID}`, { method: "DELETE" });
          setFlash(`Reply #${replyID} deleted`, false);
          return renderForumTopicDetail(id);
        } catch (err) {
          setFlash(err.message, true);
        }
      });
    });
  } catch (err) {
    app.innerHTML = `<div class="detail-card"><p>Load forum topic failed: ${escapeHTML(err.message)}</p></div>`;
  }
}
