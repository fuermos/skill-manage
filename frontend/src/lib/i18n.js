const translations = {
  zh: {
    app: { title: 'Skill Sync', subtitle: '多机器技能同步' },
    nav: { dashboard: '仪表盘', skills: '技能', sync: '同步', settings: '设置' },
    status: { connected: '已连接', offline: '离线', rev: '版本' },
    dashboard: {
      title: '仪表盘',
      totalSkills: '技能总数',
      totalSkillsDesc: '所有工具合计',
      tools: '已安装工具',
      toolsDesc: '已安装 / 已配置',
      combinations: '组合',
      combinationsDesc: '可用技能组合',
      installedTools: '已安装工具',
      recommendations: '推荐',
      noRecs: '暂无推荐，使用更多技能后生成',
      files: '个文件',
      notInstalled: '未安装'
    },
    skills: {
      title: '技能',
      search: '搜索...',
      allTools: '所有工具',
      allCategories: '所有分类',
      back: '← 返回',
      bytes: '字节'
    },
    sync: {
      title: '同步',
      push: '推送到服务器',
      pull: '从服务器拉取',
      refresh: '刷新',
      pendingChanges: '待提交变更',
      noPending: '没有待提交的变更',
      syncLog: '同步日志',
      ready: '就绪',
      pushed: '推送完成',
      pushFailed: '推送失败',
      pulled: '拉取完成',
      pullFailed: '拉取失败',
      changes: '项变更',
      applied: '已应用'
    },
    settings: {
      title: '设置',
      serverConnection: '服务器连接',
      serverUrl: '服务器地址',
      authToken: '认证令牌',
      saveTest: '保存并测试',
      saved: '已保存',
      failed: '失败',
      configuredTools: '已配置工具',
      installed: '已安装',
      notInstalled: '未安装',
      enabled: '启用',
      disabled: '禁用'
    }
  },
  en: {
    app: { title: 'Skill Sync', subtitle: 'Multi-machine Skill Sharing' },
    nav: { dashboard: 'Dashboard', skills: 'Skills', sync: 'Sync', settings: 'Settings' },
    status: { connected: 'Connected', offline: 'Offline', rev: 'Rev' },
    dashboard: {
      title: 'Dashboard',
      totalSkills: 'Skills',
      totalSkillsDesc: 'Total across all tools',
      tools: 'Tools',
      toolsDesc: 'Installed / Configured',
      combinations: 'Combos',
      combinationsDesc: 'Skill combos available',
      installedTools: 'Installed Tools',
      recommendations: 'Recommendations',
      noRecs: 'No recommendations yet. Use more skills.',
      files: 'files',
      notInstalled: 'not installed'
    },
    skills: {
      title: 'Skills',
      search: 'Search...',
      allTools: 'All Tools',
      allCategories: 'All Categories',
      back: '← Back',
      bytes: 'bytes'
    },
    sync: {
      title: 'Sync',
      push: 'Push to Server',
      pull: 'Pull from Server',
      refresh: 'Refresh',
      pendingChanges: 'Pending Changes',
      noPending: 'No pending changes',
      syncLog: 'Sync Log',
      ready: 'Ready',
      pushed: 'Push done',
      pushFailed: 'Push failed',
      pulled: 'Pull done',
      pullFailed: 'Pull failed',
      changes: 'changes',
      applied: 'Applied'
    },
    settings: {
      title: 'Settings',
      serverConnection: 'Server Connection',
      serverUrl: 'Server URL',
      authToken: 'Auth Token',
      saveTest: 'Save & Test',
      saved: 'Saved',
      failed: 'Failed',
      configuredTools: 'Configured Tools',
      installed: 'installed',
      notInstalled: 'not installed',
      enabled: 'enabled',
      disabled: 'disabled'
    }
  }
};

let currentLocale = (typeof localStorage !== 'undefined' && localStorage.getItem('skill-sync-lang')) || 'zh';

export function t(key) {
  const keys = key.split('.');
  let val = translations[currentLocale];
  for (const k of keys) {
    if (val && val[k] !== undefined) val = val[k];
    else return key;
  }
  return val;
}

export function getLocale() { return currentLocale; }
export function setLocale(locale) {
  currentLocale = locale;
  localStorage.setItem('skill-sync-lang', locale);
}