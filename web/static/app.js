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
  // roadmap ページでは nodeData を使い、index ページでは ROADMAP_NODE_IDS を使う
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
      // ノード右上に ✓ サークル
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

function openPanel(nodeId) {
  currentNodeId = nodeId;
  const data = nodeData[nodeId];
  if (!data) return;

  document.getElementById('panel-title').textContent = data.title;

  // 進捗状態 select
  const sel = document.getElementById('node-state');
  if (sel) sel.value = getNodeState(nodeId).state;

  // コンテンツ (HTML はビルド時生成済み)
  const content = document.getElementById('panel-content');
  if (content) content.innerHTML = data.html || '';

  // チェックリストを復元
  restoreChecklists(nodeId, content);

  // edit link
  const editLink = document.getElementById('edit-link');
  if (editLink && cfg.repo) {
    editLink.href = `${cfg.repo}/edit/${cfg.editBranch}/content/${nodeId}.md`;
    editLink.textContent = 'この記事を編集 (GitHub)';
  }

  panel?.classList.add('open');
  panel?.removeAttribute('aria-hidden');
  overlay?.classList.add('show');
  document.dispatchEvent(new CustomEvent('roadmapper:panelopen', { detail: { nodeId } }));

  // URL に #nodeId を設定
  history.replaceState(null, '', '#' + nodeId);
}

function closePanel() {
  panel?.classList.remove('open');
  panel?.setAttribute('aria-hidden', 'true');
  overlay?.classList.remove('show');
  currentNodeId = null;
  history.replaceState(null, '', location.pathname);
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
      // 自動遷移: 1つでもチェック → in-progress、全部 → done
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
  const savedTheme = localStorage.getItem('roadmapper:theme') || 'light';
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
