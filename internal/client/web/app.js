const API = '/api/local';

function log(msg, cls) {
  const el = document.getElementById('sync-log');
  if (el) el.innerHTML += `<div class="entry ${cls||'info'}">[${new Date().toLocaleTimeString()}] ${msg}</div>`;
}

async function fetchAPI(path) {
  try {
    const r = await fetch(API + path);
    if (!r.ok) throw new Error(r.status);
    return await r.json();
  } catch(e) { return null; }
}

async function postAPI(path, body) {
  try {
    const r = await fetch(API + path, {
      method: 'POST',
      headers: {'Content-Type':'application/json'},
      body: JSON.stringify(body)
    });
    return await r.json();
  } catch(e) { return null; }
}

document.querySelectorAll('.tab').forEach(btn => {
  btn.addEventListener('click', () => {
    document.querySelectorAll('.tab').forEach(b => b.classList.remove('active'));
    document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
    btn.classList.add('active');
    document.getElementById('tab-'+btn.dataset.tab).classList.add('active');
    refreshTab(btn.dataset.tab);
  });
});

function refreshTab(tab) {
  switch(tab) {
    case 'dashboard': loadDashboard(); break;
    case 'skills': loadSkills(); break;
    case 'sync': loadSync(); break;
    case 'settings': loadSettings(); break;
  }
}

async function checkConnection() {
  const st = await fetchAPI('/status');
  const el = document.getElementById('conn-status');
  if (st) {
    el.className = 'conn-ok'; el.textContent = 'Connected';
    document.getElementById('sync-info').textContent = 'rev ' + (st.revision||0);
  } else {
    el.className = 'conn-err'; el.textContent = 'Disconnected';
  }
}

async function loadDashboard() {
  const st = await fetchAPI('/status');
  if (st) {
    document.getElementById('stat-revision').textContent = st.revision;
    document.getElementById('stat-sync-status').textContent = 'Online';
  }

  const local = await fetchAPI('/local-status');
  if (local) {
    document.getElementById('stat-local-files').textContent = local.total;
    document.getElementById('stat-sync-status').textContent = local.synced ? 'Synced' : 'Pending';
  }

  const tools = await fetchAPI('/tools');
  const tl = document.getElementById('tool-list');
  if (tools && tools.length) {
    tl.innerHTML = tools.map(t =>
      `<div><b>${t.name}</b> ${t.installed?'<span class="conn-ok">Installed</span>':'<span class="conn-err">Not Installed</span>'} - ${t.files} files</div>`
    ).join('');
  }

  const recs = await fetchAPI('/recommendations');
  const rl = document.getElementById('rec-list');
  if (recs && recs.length) {
    rl.innerHTML = recs.slice(0,5).map(r =>
      `<div>${r.to_skill_id} <span style="color:#7c5cfc">${(r.score*100).toFixed(0)}%</span><br><small>${r.reason}</small></div>`
    ).join('');
  }
}

async function loadSkills() {
  const search = document.getElementById('skill-search').value;
  const tool = document.getElementById('skill-tool-filter').value;
  const cat = document.getElementById('skill-cat-filter').value;

  let params = [];
  if (search) params.push('search='+encodeURIComponent(search));
  if (tool) params.push('tool='+tool);
  if (cat) params.push('category='+cat);
  const qs = params.length ? '?'+params.join('&') : '';

  const skills = await fetchAPI('/skills'+qs);
  const grid = document.getElementById('skill-grid');
  if (!skills || !skills.length) { grid.innerHTML = '<p>No skills found</p>'; return; }

  grid.innerHTML = skills.map(s => `
    <div class="skill-card" onclick="showSkillDetail('${s.id}')">
      <div class="name">${s.name||s.id}</div>
      <div class="summary">${s.summary||''}</div>
      <div class="meta">
        <span>${s.tool||''}</span>
        <span>${s.category||''}</span>
        <span>${s.usage_count||0} uses</span>
        <span>${s.avg_rating ? '⭐'+s.avg_rating.toFixed(1) : ''}</span>
      </div>
      ${(s.tags||[]).map(t=>`<span class="tag">${t}</span>`).join('')}
    </div>
  `).join('');
}

async function showSkillDetail(id) {
  const skill = await fetchAPI('/skills/'+id);
  if (!skill) return;

  const grid = document.getElementById('skill-grid');
  const detail = document.getElementById('skill-detail');
  grid.classList.add('hidden');

  detail.classList.remove('hidden');
  detail.innerHTML = `
    <button class="back-btn" onclick="closeSkillDetail()">← Back</button>
    <div class="skill-detail-panel">
      <h2>${skill.display_name||skill.name}</h2>
      <p class="desc">${skill.description||skill.summary||''}</p>
      <div class="tags-bar">${(skill.tags||[]).map(t=>`<span class="tag">${t}</span>`).join('')}</div>
      <div class="stats-row">
        <div class="stat-box"><div class="num">${skill.usage_count||0}</div><div class="lbl">Uses</div></div>
        <div class="stat-box"><div class="num">${skill.avg_rating?skill.avg_rating.toFixed(1):'-'}</div><div class="lbl">Rating</div></div>
        <div class="stat-box"><div class="num">${skill.tool||'-'}</div><div class="lbl">Tool</div></div>
        <div class="stat-box"><div class="num">${skill.category||'-'}</div><div class="lbl">Category</div></div>
      </div>
    </div>
  `;
}

function closeSkillDetail() {
  document.getElementById('skill-grid').classList.remove('hidden');
  document.getElementById('skill-detail').classList.add('hidden');
}

async function loadSync() {
  const diff = await fetchAPI('/diff');
  const pc = document.getElementById('pending-changes');
  if (diff && diff.changes && diff.changes.length) {
    pc.innerHTML = `<p>${diff.changes.length} files differ</p>` + diff.changes.map(c => `<div>${c.path} [${c.status}]</div>`).join('');
  } else { pc.innerHTML = '<p>No pending changes</p>'; }
}

async function loadSettings() {
  const cfg = await fetchAPI('/config');
  if (cfg) {
    document.getElementById('cfg-server-url').value = cfg.server||'';
    document.getElementById('cfg-token').value = cfg.token||'';
  }
  const tools = await fetchAPI('/tools');
  const tl = document.getElementById('cfg-tools');
  if (tools) {
    tl.innerHTML = tools.map(t => `
      <div style="display:flex;align-items:center;gap:8px;padding:4px 0">
        <input type="checkbox" ${t.enabled?'checked':''} disabled>
        <span>${t.name}</span>
        <span class="${t.installed?'conn-ok':'conn-err'}" style="font-size:11px">${t.installed?'installed':'not installed'}</span>
      </div>
    `).join('');
  }
}

document.getElementById('btn-push').addEventListener('click', async () => {
  log('Pushing changes...');
  const r = await postAPI('/push', {});
  if (r) log(`Pushed: ${r.applied||0} changes (rev ${r.new_revision})`, 'success');
  else log('Push failed', 'error');
  loadSync();
});

document.getElementById('btn-pull').addEventListener('click', async () => {
  log('Pulling from server...');
  const r = await postAPI('/pull', {});
  if (r) log(`Pulled: ${r.changes?.length||0} changes`, 'success');
  else log('Pull failed', 'error');
  loadSync();
});

document.getElementById('btn-diff').addEventListener('click', loadSync);

document.getElementById('skill-search').addEventListener('input', loadSkills);
document.getElementById('skill-tool-filter').addEventListener('change', loadSkills);
document.getElementById('skill-cat-filter').addEventListener('change', loadSkills);

document.getElementById('cfg-save').addEventListener('click', async () => {
  const r = await postAPI('/config/save', {
    server: document.getElementById('cfg-server-url').value,
    token: document.getElementById('cfg-token').value,
  });
  document.getElementById('cfg-test-result').textContent = r ? 'Saved' : 'Failed';
  setTimeout(checkConnection, 500);
});

checkConnection();
loadDashboard();
setInterval(checkConnection, 30000);