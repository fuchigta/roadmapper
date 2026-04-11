/* roadmapper app.js — サイドパネル / 進捗トラッキング / テーマ */

'use strict';

// ===== Storage =====
const STORAGE_KEY = 'roadmapper:progress';

function loadProgress() {
  try { return JSON.parse(localStorage.getItem(STORAGE_KEY) || '{}'); }
  catch { return {}; }
}

function saveProgress(data) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(data));
}

// ===== State =====
let progress = loadProgress();
const cfg = window.SITE_CONFIG || {};
const nodeData = window.ROADMAP_DATA || {};
const roadmapId = cfg.roadmapId || '';

function getNodeState(nodeId) {
  return (progress[roadmapId]?.[nodeId]) || { state: 'none', tasks: [] };
}

function setNodeState(nodeId, patch) {
  if (!progress[roadmapId]) progress[roadmapId] = {};
  const cur = getNodeState(nodeId);
  progress[roadmapId][nodeId] = { ...cur, ...patch };
  saveProgress(progress);
}

// ===== Progress calculation =====
function calcRoadmapProgress(rmId) {
  const rm = progress[rmId] || {};
  const nodeIds = (window.ROADMAP_NODE_IDS?.[rmId]) || Object.keys(nodeData);
  if (!nodeIds.length) return 0;
  const done = nodeIds.filter(id => rm[id]?.state === 'done').length;
  return Math.round((done / nodeIds.length) * 100);
}

function updateProgressBar() {
  const pct = calcRoadmapProgress(roadmapId);
  const bar = document.getElementById('progress-bar');
  const label = document.getElementById('progress-label');
  if (bar) bar.style.setProperty('--progress', pct + '%');
  if (label) label.textContent = pct + '%';
}

function updateNodeVisuals() {
  document.querySelectorAll('.roadmap-node[data-id]').forEach(el => {
    const id = el.dataset.id;
    const { state, tasks } = getNodeState(id);
    el.dataset.state = state;

    // done バッジの表示
    let indicator = el.querySelector('.node-indicator');
    if (state === 'done') {
      if (!indicator) {
        indicator = document.createElementNS('http://www.w3.org/2000/svg', 'g');
        indicator.setAttribute('class', 'node-indicator');
        el.appendChild(indicator);
      }
      const rect = el.querySelector('rect');
      if (rect) {
        const rx = parseFloat(rect.getAttribute('x') || 0) + parseFloat(rect.getAttribute('width') || 0);
        const ry = parseFloat(rect.getAttribute('y') || 0);
        indicator.innerHTML =
          `<circle cx="${rx}" cy="${ry}" r="9" fill="#16a34a" stroke="white" stroke-width="1.5"/>` +
          `<text x="${rx}" y="${ry}" text-anchor="middle" dominant-baseline="middle" ` +
          `font-family="system-ui" font-size="10" fill="white" font-weight="bold">✓</text>`;
      }
    } else if (indicator) {
      indicator.innerHTML = '';
    }

    // チェックリスト進捗をノードタイトル下に小さく表示
    const nodeMeta = nodeData[id];
    if (nodeMeta?.html) {
      const totalTasks = (nodeMeta.html.match(/type="checkbox"/g) || []).length;
      if (totalTasks > 0) {
        const checkedTasks = (tasks || []).filter(Boolean).length;
        let taskBadge = el.querySelector('.task-badge');
        if (!taskBadge) {
          taskBadge = document.createElementNS('http://www.w3.org/2000/svg', 'text');
          taskBadge.setAttribute('class', 'task-badge');
          taskBadge.setAttribute('text-anchor', 'middle');
          taskBadge.setAttribute('dominant-baseline', 'middle');
          taskBadge.setAttribute('font-family', 'system-ui,sans-serif');
          taskBadge.setAttribute('font-size', '9');
          taskBadge.setAttribute('fill', '#94a3b8');
          el.appendChild(taskBadge);
        }
        const rect = el.querySelector('rect');
        if (rect) {
          const cy = parseFloat(rect.getAttribute('y') || 0) + parseFloat(rect.getAttribute('height') || 0) - 8;
          const cx = parseFloat(rect.getAttribute('x') || 0) + parseFloat(rect.getAttribute('width') || 0) / 2;
          taskBadge.setAttribute('x', cx);
          taskBadge.setAttribute('y', cy);
          taskBadge.textContent = `${checkedTasks}/${totalTasks}`;
        }
      }
    }
  });
}

// ===== Side panel =====
const panel = document.getElementById('side-panel');
const overlay = document.getElementById('overlay');
let currentNodeId = null;

// ===== Panel resize =====
const PANEL_WIDTH_KEY = 'roadmapper:panel-width';
const PANEL_MIN = 320;
function panelMax() { return Math.min(window.innerWidth * 0.8, 960); }

function applySavedPanelWidth() {
  if (!panel || window.innerWidth <= 600) return;
  const saved = parseInt(localStorage.getItem(PANEL_WIDTH_KEY) || '', 10);
  if (!isNaN(saved)) {
    panel.style.width = Math.max(PANEL_MIN, Math.min(saved, panelMax())) + 'px';
  }
}

function initPanelResize() {
  const resizer = document.getElementById('panel-resizer');
  if (!resizer || !panel) return;
  let startX = 0;
  let startWidth = 0;

  resizer.addEventListener('pointerdown', (e) => {
    if (window.innerWidth <= 600) return;
    e.preventDefault();
    startX = e.clientX;
    startWidth = panel.offsetWidth;
    resizer.classList.add('dragging');
    panel.classList.add('resizing');
    resizer.setPointerCapture(e.pointerId);
  });

  resizer.addEventListener('pointermove', (e) => {
    if (!resizer.classList.contains('dragging')) return;
    const delta = startX - e.clientX;
    const newWidth = Math.max(PANEL_MIN, Math.min(startWidth + delta, panelMax()));
    panel.style.width = newWidth + 'px';
  });

  resizer.addEventListener('pointerup', () => {
    if (!resizer.classList.contains('dragging')) return;
    resizer.classList.remove('dragging');
    panel.classList.remove('resizing');
    localStorage.setItem(PANEL_WIDTH_KEY, parseInt(panel.style.width) || '');
  });
}

applySavedPanelWidth();
initPanelResize();

// ===== 依存関係ハイライト =====
function highlightRelations(nodeId) {
  const data = nodeData[nodeId];
  if (!data) return;

  const relatedIds = new Set([nodeId, ...(data.parents || []), ...(data.children || [])]);

  document.querySelectorAll('.roadmap-node[data-id]').forEach(el => {
    if (relatedIds.has(el.dataset.id)) {
      el.classList.remove('dim');
      el.classList.toggle('highlight', el.dataset.id !== nodeId);
    } else {
      el.classList.add('dim');
      el.classList.remove('highlight');
    }
  });

  document.querySelectorAll('.roadmap-edge').forEach(el => {
    const src = el.dataset.source;
    const tgt = el.dataset.target;
    const related = relatedIds.has(src) && relatedIds.has(tgt);
    el.classList.toggle('dim', !related);
    el.classList.toggle('highlight', related && (src === nodeId || tgt === nodeId));
  });
}

function clearHighlights() {
  document.querySelectorAll('.roadmap-node').forEach(el => {
    el.classList.remove('dim', 'highlight');
  });
  document.querySelectorAll('.roadmap-edge').forEach(el => {
    el.classList.remove('dim', 'highlight');
  });
}

function renderRelations(nodeId) {
  const data = nodeData[nodeId];
  const relationsEl = document.getElementById('panel-relations');
  const parentsEl = document.getElementById('panel-parents');
  const childrenEl = document.getElementById('panel-children');
  const parentsList = document.getElementById('panel-parents-list');
  const childrenList = document.getElementById('panel-children-list');
  if (!relationsEl || !data) return;

  function makeItems(ids, listEl) {
    listEl.innerHTML = '';
    ids.forEach(id => {
      const nd = nodeData[id];
      if (!nd) return;
      const li = document.createElement('li');
      const a = document.createElement('a');
      a.href = '#' + id;
      a.textContent = nd.title;
      a.addEventListener('click', (e) => { e.preventDefault(); openPanel(id); });
      li.appendChild(a);
      listEl.appendChild(li);
    });
  }

  const hasParents = (data.parents || []).length > 0;
  const hasChildren = (data.children || []).length > 0;

  parentsEl.hidden = !hasParents;
  childrenEl.hidden = !hasChildren;
  relationsEl.hidden = !hasParents && !hasChildren;

  makeItems(data.parents || [], parentsList);
  makeItems(data.children || [], childrenList);
}

function openPanel(nodeId) {
  currentNodeId = nodeId;
  const data = nodeData[nodeId];
  if (!data) return;

  document.getElementById('panel-title').textContent = data.title;

  const sel = document.getElementById('node-state');
  if (sel) sel.value = getNodeState(nodeId).state;

  const content = document.getElementById('panel-content');
  if (content) content.innerHTML = data.html || '';

  restoreChecklists(nodeId, content);

  renderRelations(nodeId);
  highlightRelations(nodeId);

  const editLink = document.getElementById('edit-link');
  if (editLink && cfg.repo) {
    editLink.href = `${cfg.repo}/edit/${cfg.editBranch}/content/${nodeId}.md`;
    editLink.textContent = 'この記事を編集 (GitHub)';
  }

  panel?.classList.add('open');
  panel?.removeAttribute('aria-hidden');
  overlay?.classList.add('show');
  document.dispatchEvent(new CustomEvent('roadmapper:panelopen', { detail: { nodeId } }));

  history.replaceState(null, '', '#' + nodeId);
}

function closePanel() {
  panel?.classList.remove('open');
  panel?.setAttribute('aria-hidden', 'true');
  overlay?.classList.remove('show');
  currentNodeId = null;
  clearHighlights();
  history.replaceState(null, '', location.pathname + location.search);
}

// ===== チェックリスト =====
function restoreChecklists(nodeId, container) {
  if (!container) return;
  const { tasks } = getNodeState(nodeId);
  container.querySelectorAll('input[type="checkbox"]').forEach((cb, i) => {
    cb.checked = !!tasks[i];
    cb.addEventListener('change', () => {
      const st = getNodeState(nodeId);
      const newTasks = [...(st.tasks || [])];
      newTasks[i] = cb.checked;
      const all = container.querySelectorAll('input[type="checkbox"]');
      const checkedCount = [...all].filter(c => c.checked).length;
      let newState = st.state;
      if (checkedCount === 0) newState = 'none';
      else if (checkedCount === all.length) newState = 'done';
      else newState = 'in-progress';
      setNodeState(nodeId, { tasks: newTasks, state: newState });
      const sel = document.getElementById('node-state');
      if (sel) sel.value = newState;
      updateProgressBar();
      updateNodeVisuals();
    });
  });
}

// ===== ノード検索 =====
function initSearch() {
  const input = document.getElementById('node-search');
  const results = document.getElementById('search-results');
  if (!input || !results) return;

  const allNodes = Object.entries(nodeData).map(([id, d]) => ({ id, title: d.title }));
  let selectedIdx = -1;

  function showResults(query) {
    results.innerHTML = '';
    selectedIdx = -1;
    if (!query) { results.hidden = true; return; }

    const q = query.toLowerCase();
    const matched = allNodes.filter(n => n.title.toLowerCase().includes(q));
    if (!matched.length) { results.hidden = true; return; }

    matched.slice(0, 8).forEach((n, i) => {
      const li = document.createElement('li');
      li.textContent = n.title;
      li.dataset.idx = i;
      li.addEventListener('mousedown', (e) => {
        e.preventDefault();
        selectResult(n.id);
      });
      results.appendChild(li);
    });
    results.hidden = false;
  }

  function selectResult(nodeId) {
    input.value = '';
    results.hidden = true;

    // SVG ノードへスクロール
    const el = document.querySelector(`.roadmap-node[data-id="${nodeId}"]`);
    if (el) {
      el.classList.add('search-highlight');
      el.scrollIntoView({ behavior: 'smooth', block: 'center' });
      el.addEventListener('animationend', () => el.classList.remove('search-highlight'), { once: true });
    }
    openPanel(nodeId);
  }

  input.addEventListener('input', () => showResults(input.value));

  input.addEventListener('keydown', (e) => {
    const items = results.querySelectorAll('li');
    if (!items.length) return;

    if (e.key === 'ArrowDown') {
      e.preventDefault();
      selectedIdx = Math.min(selectedIdx + 1, items.length - 1);
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      selectedIdx = Math.max(selectedIdx - 1, 0);
    } else if (e.key === 'Enter' && selectedIdx >= 0) {
      e.preventDefault();
      const nodeId = Object.keys(nodeData).filter(id =>
        nodeData[id].title.toLowerCase().includes(input.value.toLowerCase())
      )[selectedIdx];
      if (nodeId) selectResult(nodeId);
      return;
    } else if (e.key === 'Escape') {
      results.hidden = true;
      input.blur();
      return;
    } else {
      return;
    }

    items.forEach((li, i) => li.classList.toggle('selected', i === selectedIdx));
  });

  document.addEventListener('click', (e) => {
    if (!input.contains(e.target) && !results.contains(e.target)) {
      results.hidden = true;
    }
  });
}

// ===== キーボードナビゲーション =====
function initKeyboard() {
  // go 側の json.Marshal はキーをアルファベット順にするため、DAG 順序を __order で渡している
  const nodeIds = nodeData.__order || Object.keys(nodeData).filter(k => k !== '__order');

  document.addEventListener('keydown', (e) => {
    // 入力フォーカス中は無視
    const tag = document.activeElement?.tagName?.toLowerCase();
    if (tag === 'input' || tag === 'select' || tag === 'textarea') return;

    // / キーで検索フォーカス
    if (e.key === '/') {
      e.preventDefault();
      document.getElementById('node-search')?.focus();
      return;
    }

    // Escape でパネルを閉じる
    if (e.key === 'Escape' && currentNodeId) {
      closePanel();
      return;
    }

    // j/k または ArrowDown/ArrowUp でノード移動
    if (e.key === 'j' || e.key === 'ArrowDown' || e.key === 'k' || e.key === 'ArrowUp') {
      e.preventDefault();
      const dir = (e.key === 'j' || e.key === 'ArrowDown') ? 1 : -1;
      // パネルが閉じている場合は j/↓ のみ先頭ノードを開く
      if (!currentNodeId) {
        if (dir > 0 && nodeIds[0]) openPanel(nodeIds[0]);
        return;
      }
      const idx = nodeIds.indexOf(currentNodeId);
      const nextIdx = idx + dir;
      // 境界外なら何もしない
      if (nextIdx < 0 || nextIdx >= nodeIds.length) return;
      openPanel(nodeIds[nextIdx]);
      return;
    }

    // 1-4 で進捗状態切替 (パネル表示中のみ)
    if (currentNodeId && ['1', '2', '3', '4'].includes(e.key)) {
      const states = ['none', 'in-progress', 'done', 'skipped'];
      const newState = states[parseInt(e.key) - 1];
      setNodeState(currentNodeId, { state: newState });
      const sel = document.getElementById('node-state');
      if (sel) sel.value = newState;
      updateProgressBar();
      updateNodeVisuals();
    }
  });
}

// ===== SVG ズーム・パン =====
function initZoomPan() {
  const wrap = document.getElementById('diagram-wrap');
  const svg = wrap?.querySelector('.roadmap-svg');
  if (!wrap || !svg) return;

  let scale = 1;
  let tx = 0;
  let ty = 0;
  const MIN_SCALE = 0.2;
  const MAX_SCALE = 3;

  function applyTransform() {
    svg.style.transform = `translate(${tx}px, ${ty}px) scale(${scale})`;
  }

  function clampTranslate() {
    // 中心付近に留まるよう緩やかに制限
  }

  // ホイールズーム
  wrap.addEventListener('wheel', (e) => {
    e.preventDefault();
    const rect = wrap.getBoundingClientRect();
    const mx = e.clientX - rect.left;
    const my = e.clientY - rect.top;

    const delta = e.deltaY < 0 ? 1.1 : 0.9;
    const newScale = Math.max(MIN_SCALE, Math.min(MAX_SCALE, scale * delta));

    // ポインタ位置を中心にズーム
    tx = mx - (mx - tx) * (newScale / scale);
    ty = my - (my - ty) * (newScale / scale);
    scale = newScale;
    applyTransform();
  }, { passive: false });

  // ピンチズーム (タッチ)
  let lastDist = 0;
  let lastMidX = 0;
  let lastMidY = 0;

  wrap.addEventListener('touchstart', (e) => {
    if (e.touches.length === 2) {
      const t0 = e.touches[0];
      const t1 = e.touches[1];
      lastDist = Math.hypot(t1.clientX - t0.clientX, t1.clientY - t0.clientY);
      lastMidX = (t0.clientX + t1.clientX) / 2;
      lastMidY = (t0.clientY + t1.clientY) / 2;
    }
  }, { passive: true });

  wrap.addEventListener('touchmove', (e) => {
    if (e.touches.length === 2) {
      e.preventDefault();
      const t0 = e.touches[0];
      const t1 = e.touches[1];
      const dist = Math.hypot(t1.clientX - t0.clientX, t1.clientY - t0.clientY);
      const midX = (t0.clientX + t1.clientX) / 2;
      const midY = (t0.clientY + t1.clientY) / 2;

      const rect = wrap.getBoundingClientRect();
      const mx = midX - rect.left;
      const my = midY - rect.top;

      const delta = dist / lastDist;
      const newScale = Math.max(MIN_SCALE, Math.min(MAX_SCALE, scale * delta));
      tx = mx - (mx - tx) * (newScale / scale);
      ty = my - (my - ty) * (newScale / scale);
      scale = newScale;

      lastDist = dist;
      lastMidX = midX;
      lastMidY = midY;
      applyTransform();
    }
  }, { passive: false });

  // ドラッグパン
  let dragging = false;
  let dragStartX = 0;
  let dragStartY = 0;
  let dragTx = 0;
  let dragTy = 0;

  wrap.addEventListener('pointerdown', (e) => {
    if (e.target.closest('.roadmap-node') || e.target.closest('.zoom-reset')) return;
    dragging = true;
    dragStartX = e.clientX;
    dragStartY = e.clientY;
    dragTx = tx;
    dragTy = ty;
    wrap.classList.add('panning');
    wrap.setPointerCapture(e.pointerId);
  });

  wrap.addEventListener('pointermove', (e) => {
    if (!dragging) return;
    tx = dragTx + (e.clientX - dragStartX);
    ty = dragTy + (e.clientY - dragStartY);
    applyTransform();
  });

  wrap.addEventListener('pointerup', () => {
    dragging = false;
    wrap.classList.remove('panning');
  });

  // リセット
  document.getElementById('zoom-reset')?.addEventListener('click', () => {
    scale = 1;
    tx = 0;
    ty = 0;
    applyTransform();
  });
}

// ===== 進捗 URL シェア =====
function encodeProgress(data) {
  try {
    return btoa(unescape(encodeURIComponent(JSON.stringify(data))));
  } catch { return ''; }
}

function decodeProgress(str) {
  try {
    return JSON.parse(decodeURIComponent(escape(atob(str))));
  } catch { return null; }
}

function showToast(msg) {
  const el = document.createElement('div');
  el.className = 'share-toast';
  el.textContent = msg;
  document.body.appendChild(el);
  setTimeout(() => {
    el.classList.add('fade-out');
    el.addEventListener('transitionend', () => el.remove());
  }, 2000);
}

function initShare() {
  // シェアボタン
  document.getElementById('share-btn')?.addEventListener('click', () => {
    const encoded = encodeProgress(progress);
    if (!encoded) return;
    const url = new URL(location.href);
    url.searchParams.set('p', encoded);
    url.hash = '';
    navigator.clipboard.writeText(url.toString()).then(() => {
      showToast('進捗 URL をクリップボードにコピーしました');
    }).catch(() => {
      showToast(url.toString());
    });
  });

  // シェア URL からの読み込み (読み取り専用ビュー)
  const urlParams = new URLSearchParams(location.search);
  const sharedParam = urlParams.get('p');
  if (sharedParam) {
    const shared = decodeProgress(sharedParam);
    if (shared) {
      // シェアビューであることを通知
      progress = shared;
      showToast('シェアされた進捗を表示中（読み取り専用）');
      // シェアビューでは localStorage に書き込まない
    }
  }
}

// ===== Event wiring =====
document.addEventListener('DOMContentLoaded', () => {
  // ノードクリック
  document.querySelectorAll('.roadmap-node[data-id]').forEach(el => {
    el.addEventListener('click', () => openPanel(el.dataset.id));
  });

  // パネル閉じる
  document.getElementById('panel-close')?.addEventListener('click', closePanel);
  overlay?.addEventListener('click', closePanel);

  // 進捗 select
  document.getElementById('node-state')?.addEventListener('change', e => {
    if (!currentNodeId) return;
    setNodeState(currentNodeId, { state: e.target.value });
    updateProgressBar();
    updateNodeVisuals();
  });

  // テーマトグル
  const toggleBtn = document.getElementById('theme-toggle');
  const savedTheme = localStorage.getItem('roadmapper:theme') ||
    (window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light');
  document.documentElement.dataset.theme = savedTheme;
  updateThemeIcon(savedTheme);
  toggleBtn?.addEventListener('click', () => {
    const next = document.documentElement.dataset.theme === 'dark' ? 'light' : 'dark';
    document.documentElement.dataset.theme = next;
    localStorage.setItem('roadmapper:theme', next);
    updateThemeIcon(next);
  });

  // エクスポート
  document.getElementById('export-btn')?.addEventListener('click', () => {
    const blob = new Blob([JSON.stringify(progress, null, 2)], { type: 'application/json' });
    const a = document.createElement('a');
    a.href = URL.createObjectURL(blob);
    a.download = 'roadmapper-progress.json';
    a.click();
  });

  // インポート
  const importBtn = document.getElementById('import-btn');
  const importFile = document.getElementById('import-file');
  importBtn?.addEventListener('click', () => importFile?.click());
  importFile?.addEventListener('change', e => {
    const file = e.target.files?.[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = evt => {
      try {
        const imported = JSON.parse(evt.target.result);
        progress = { ...progress, ...imported };
        saveProgress(progress);
        updateProgressBar();
        updateNodeVisuals();
      } catch {
        alert('ファイルの読み込みに失敗しました');
      }
    };
    reader.readAsText(file);
    e.target.value = '';
  });

  // 各機能の初期化
  initSearch();
  initKeyboard();
  initZoomPan();
  initShare();

  // URL の #nodeId でパネルを開く
  if (location.hash) {
    const id = location.hash.slice(1);
    if (nodeData[id]) openPanel(id);
  }

  // index ページのカード進捗
  const roadmapIds = window.ROADMAP_IDS || [];
  roadmapIds.forEach(rmId => {
    const pct = calcRoadmapProgress(rmId);
    const card = document.querySelector(`.card-progress[data-roadmap="${rmId}"]`);
    if (!card) return;
    const bar = card.querySelector('.progress-bar');
    const label = card.querySelector('.progress-label');
    if (bar) bar.style.setProperty('--progress', pct + '%');
    if (label) label.textContent = pct + '%';
  });

  updateProgressBar();
  updateNodeVisuals();
});

function updateThemeIcon(theme) {
  const btn = document.getElementById('theme-toggle');
  if (btn) btn.textContent = theme === 'dark' ? '☀️' : '🌙';
}
