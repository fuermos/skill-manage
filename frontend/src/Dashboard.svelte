<script>
  import { onMount } from 'svelte';
  import { t as _ } from './lib/i18n.js';

  let tools = [];
  let recs = [];

  async function load() {
    try {
      tools = await window.go.main.App.GetTools();
      recs = await window.go.main.App.GetRecommendations();
    } catch(e) {}
  }

  onMount(() => { load(); });
</script>

<div>
  <h2 style="margin-bottom:20px;font-size:20px">{_('dashboard.title')}</h2>

  <div class="grid-3">
    <div class="card">
      <div class="card-title">{_('dashboard.totalSkills')}</div>
      <div class="card-val">{tools.reduce((s,t)=>s+(t.files||0), 0)}</div>
      <small>{_('dashboard.totalSkillsDesc')}</small>
    </div>
    <div class="card">
      <div class="card-title">{_('dashboard.tools')}</div>
      <div class="card-val">{tools.filter(t=>t.installed).length}/{tools.length}</div>
      <small>{_('dashboard.toolsDesc')}</small>
    </div>
    <div class="card">
      <div class="card-title">{_('dashboard.combinations')}</div>
      <div class="card-val">--</div>
      <small>{_('dashboard.combinationsDesc')}</small>
    </div>
  </div>

  <div class="grid-2" style="margin-top:20px">
    <div class="card">
      <div class="card-title">{_('dashboard.installedTools')}</div>
      {#each tools as tool}
        <div class="tool-row">
          <span>{tool.display || tool.name}</span>
          <span class:tag-green={tool.installed} class:tag-red={!tool.installed}>
            {tool.installed ? `${tool.files} ${_('dashboard.files')}` : _('dashboard.notInstalled')}
          </span>
        </div>
      {/each}
    </div>
    <div class="card">
      <div class="card-title">{_('dashboard.recommendations')}</div>
      {#if recs.length}
        {#each recs.slice(0,5) as r}
          <div class="rec-row">
            <span>{r.to_skill_id}</span>
            <span class="score">{(r.score*100).toFixed(0)}%</span>
            <br><small>{r.reason}</small>
          </div>
        {/each}
      {:else}
        <p style="color:#555;font-size:13px">{_('dashboard.noRecs')}</p>
      {/if}
    </div>
  </div>
</div>

<style>
  .grid-3 { display:grid; grid-template-columns:repeat(3,1fr); gap:16px; }
  .grid-2 { display:grid; grid-template-columns:repeat(2,1fr); gap:16px; }
  .card { background:#1a1a28; border:1px solid #2a2a3a; border-radius:8px; padding:16px; }
  .card-title { font-size:12px; color:#888; text-transform:uppercase; letter-spacing:.5px; margin-bottom:8px; }
  .card-val { font-size:32px; font-weight:700; color:#7c5cfc; }
  .card small { font-size:11px; color:#555; }
  .tool-row { display:flex; justify-content:space-between; padding:6px 0; font-size:13px; border-bottom:1px solid #1e1e2e; }
  .tag-green { color:#4ade80; font-size:11px; }
  .tag-red { color:#f87171; font-size:11px; }
  .rec-row { padding:6px 0; font-size:13px; border-bottom:1px solid #1e1e2e; }
  .rec-row small { color:#555; font-size:11px; }
  .score { color:#7c5cfc; font-weight:600; float:right; }
</style>