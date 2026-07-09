export default {
  layout: {
    settings: '布局设置',
    showTagsView: '显示 TagsView',
    showBreadcrumb: '显示面包屑',
    showLogo: '显示 Logo',
    showFooter: '显示页脚',
    theme: '主题',
    primaryColor: '主题色',
    componentSize: '组件大小',
    locale: '语言'
  },
  size: {
    large: '大',
    default: '默认',
    small: '小'
  },
  common: {
    welcome: '欢迎, {name}',
    action: {
      search: '搜索',
      reset: '重置',
      create: '新增',
      edit: '编辑',
      view: '查看',
      delete: '删除',
      batchDelete: '批量删除',
      export: '导出',
      refresh: '刷新',
      confirm: '确认',
      cancel: '取消',
      close: '关闭',
      save: '保存',
      submit: '提交'
    },
    status: {
      enable: '启用',
      disable: '禁用',
      yes: '是',
      no: '否',
      success: '成功',
      failed: '失败',
      error: '错误',
      pending: '待处理',
      processing: '处理中',
      completed: '已完成',
      cancelled: '已取消',
      deleted: '已删除'
    },
    table: {
      columnSettings: '列设置',
      visibleFields: '显示字段',
      showAll: '全部显示',
      noData: '暂无数据'
    },
    placeholder: {
      input: '请输入{label}',
      select: '请选择{label}',
      selectRole: '请选择角色',
      selectUser: '请输入姓名/账号',
      selectDept: '请选择部门'
    },
    message: {
      createSuccess: '新增成功',
      updateSuccess: '更新成功',
      deleteSuccess: '删除成功',
      deleteCancelled: '已取消删除',
      createFailed: '新增失败',
      updateFailed: '更新失败',
      deleteFailed: '删除失败',
      exportSuccess: '导出成功',
      exportFailed: '导出失败',
      fieldRequired: '字段不能为空'
    },
    column: {
      status: '状态',
      description: '描述',
      createTime: '创建时间',
      updateTime: '更新时间',
      lastLoginTime: '最后登录',
      actions: '操作',
      sort: '排序',
      leader: '负责人',
      publisher: '发布人',
      publishTime: '发布时间'
    },
    notFound: {
      subTitle: '抱歉，您访问的页面不存在',
      backHome: '返回首页'
    }
  },
  auth: {
    title: '用户登录',
    username: '用户名',
    password: '密码',
    submit: '登录',
    loginFailed: '登录失败，请稍后重试',
    usernameLength: '用户名长度应在 3-20 个字符之间',
    passwordLength: '密码长度至少 6 个字符'
  },
  user: {
    title: '用户管理',
    searchKeyword: '用户名、姓名、邮箱或电话',
    addTitle: '新增用户',
    editTitle: '编辑用户',
    viewTitle: '查看用户',
    field: {
      username: '用户名',
      realName: '姓名',
      email: '邮箱',
      phone: '电话',
      roles: '角色',
      status: '状态',
      password: '密码',
      confirmPassword: '确认密码'
    },
    placeholder: {
      selectRole: '请选择角色',
      passwordAdd: '请输入密码',
      passwordEdit: '编辑时留空表示不修改'
    },
    validation: {
      usernameRequired: '请输入用户名',
      usernameLength: '用户名长度应在 3-20 个字符之间',
      realNameRequired: '请输入姓名',
      realNameLength: '姓名长度应在 2-20 个字符之间',
      emailRequired: '请输入邮箱',
      emailInvalid: '请输入有效的邮箱地址',
      phoneRequired: '请输入电话',
      phoneInvalid: '请输入有效的手机号',
      rolesRequired: '请选择角色',
      rolesAtLeastOne: '请至少选择一个角色',
      statusRequired: '请选择状态',
      passwordRequired: '请输入密码',
      passwordLength: '密码长度至少 6 个字符',
      confirmPasswordRequired: '请再次输入密码',
      confirmMismatch: '两次输入的密码不一致'
    }
  },
  role: {
    title: '角色管理',
    searchKeyword: '角色名称、代码或描述',
    permissionLabel: '权限',
    permissionConfigTitle: '权限配置',
    permissionSaveSuccess: '权限保存成功',
    addTitle: '新增角色',
    editTitle: '编辑角色',
    viewTitle: '查看角色',
    field: {
      name: '角色名称',
      code: '角色代码',
      description: '描述',
      status: '状态'
    },
    validation: {
      nameRequired: '请输入角色名称',
      nameLength: '角色名称长度应在 2-20 个字符之间',
      codeRequired: '请输入角色代码',
      codeLength: '角色代码长度应在 2-20 个字符之间',
      descriptionMax: '描述长度不能超过 200 个字符',
      statusRequired: '请选择状态'
    }
  },
  dept: {
    title: '部门管理',
    searchKeyword: '部门名称',
    addSubLabel: '新增子部门',
    deleteConfirm: '确定要删除该部门吗？删除后子部门也会被删除。',
    addTitle: '新增部门',
    editTitle: '编辑部门',
    field: {
      parentId: '上级部门',
      name: '部门名称',
      leader: '负责人',
      phone: '联系电话',
      email: '邮箱',
      sort: '排序',
      status: '状态'
    },
    placeholder: {
      noParent: '无上级部门'
    },
    validation: {
      nameRequired: '请输入部门名称'
    }
  },
  notice: {
    title: '公告管理',
    searchKeyword: '公告标题',
    priority: { high: '紧急', medium: '重要', low: '普通' },
    bannerTitle: '{priority} 公告：{title}',
    actionPublish: '发布',
    actionRevoke: '撤销',
    center: {
      title: '通知中心',
      markAllRead: '全部已读',
      viewAll: '查看全部',
      publisher: '发布人',
      publishTime: '发布时间',
      emptyAnnouncement: '暂无公告',
      emptyNotice: '暂无通知',
      emptyTodo: '暂无待办'
    },
    confirmPublish: '确定发布该公告吗？',
    confirmRevoke: '确定撤销该公告吗？',
    publishSuccess: '发布成功',
    revokeSuccess: '撤销成功',
    addTitle: '新增公告',
    editTitle: '编辑公告',
    viewTitle: '查看公告',
    field: {
      title: '标题',
      type: '类型',
      status: '状态',
      priority: '优先级',
      content: '内容',
      publishTime: '发布时间',
      expireTime: '过期时间'
    },
    option: {
      typeAnnouncement: '公告',
      typeNotice: '通知',
      typeTodo: '待办',
      statusDraft: '草稿',
      statusPublished: '已发布',
      statusExpired: '已过期',
      statusRevoked: '已撤销',
      priorityHigh: '高',
      priorityMedium: '中',
      priorityLow: '低'
    },
    validation: {
      titleRequired: '请输入标题',
      contentRequired: '请输入内容'
    }
  },
  permission: {
    title: '权限管理',
    searchKeyword: '权限名称、代码或描述',
    addTitle: '新增权限',
    editTitle: '编辑权限',
    viewTitle: '查看权限',
    field: {
      name: '权限名称',
      code: '权限代码',
      module: '模块',
      status: '状态',
      description: '描述'
    },
    option: {
      moduleSystem: '系统管理',
      moduleUser: '用户管理',
      moduleRole: '角色管理',
      modulePermission: '权限管理',
      moduleDict: '字典管理',
      moduleConfig: '系统配置'
    },
    validation: {
      nameRequired: '请输入权限名称',
      codeRequired: '请输入权限代码',
      moduleRequired: '请选择模块',
      statusRequired: '请选择状态',
      descriptionMax: '描述长度不能超过 200 个字符'
    }
  },
  menu: {
    title: '菜单管理',
    addTitle: '新增顶级菜单',
    addSubTitle: '新增子菜单 - {name}',
    editTitle: '编辑菜单',
    viewTitle: '查看菜单',
    field: {
      parentId: '父菜单',
      name: '菜单名称',
      path: '路由路径',
      component: '组件路径',
      icon: '图标',
      sort: '排序',
      status: '状态'
    },
    placeholder: {
      noParent: '不选则为顶级菜单',
      component: '如 dashboard/views/Home',
      icon: 'Element Plus 图标名'
    },
    validation: {
      nameRequired: '请输入菜单名称',
      pathRequired: '请输入路由路径',
      statusRequired: '请选择状态'
    },
    actionAddSub: '新增子菜单',
    loadFailed: '加载菜单树失败',
    deleteConfirm: '确认删除菜单「{name}」？子菜单将一并删除。',
    sortUpdated: '排序已更新'
  },
  crud: {
    title: 'CRUD 示例',
    addTitle: '新增记录',
    editTitle: '编辑记录',
    viewTitle: '查看记录',
    field: {
      name: '姓名',
      province: '省份',
      city: '城市',
      date: '日期',
      address: '地址',
      zip: '邮编'
    },
    validation: {
      nameRequired: '请输入姓名',
      provinceRequired: '请选择省份',
      cityRequired: '请输入城市',
      addressRequired: '请输入地址',
      zipRequired: '请输入邮编',
      zipPattern: '请输入6位邮编'
    }
  },
  dashboard: {
    defaultUser: '用户',
    roleSuperAdmin: '超级管理员',
    roleNormal: '普通用户',
    stat: {
      users: '总用户数',
      orders: '订单数',
      revenue: '总营收',
      active: '活跃用户'
    },
    quickAction: {
      user: '用户管理',
      role: '角色管理',
      permission: '权限管理',
      dict: '字典管理',
      menu: '菜单管理',
      notice: '公告管理',
      crud: '增删改查',
      profile: '个人中心',
      docs: '使用文档'
    },
    message: {
      inDevelopment: '该功能正在开发中',
      docsBuilding: '文档站点建设中',
      loadFailed: '首页数据加载失败'
    },
    trendLabel: '较昨日',
    quickEntry: '快捷入口',
    noQuickAction: '暂无可用快捷入口',
    latestActivity: '最新动态',
    noActivity: '暂无动态',
    visitTrend: '访问趋势',
    roleDistribution: '用户角色分布',
    lastLoginPrefix: '上次登录'
  },
  profile: {
    title: '个人中心',
    editProfile: '编辑资料',
    basicInfo: '基本信息',
    personalSettings: '个人设置',
    field: {
      username: '用户名',
      email: '邮箱',
      phone: '手机号',
      gender: '性别',
      department: '部门',
      position: '职位',
      joinDate: '入职时间',
      lastLogin: '最后登录',
      name: '姓名',
      language: '语言',
      timezone: '时区',
      notification: '通知设置',
      emailReminder: '邮件提醒'
    },
    option: {
      chinese: '中文',
      english: '英文'
    },
    saveSuccess: '资料保存成功'
  },
  dict: {
    selectNode: '请选择一个节点',
    categoryDetail: '字典分类详情',
    dictDetail: '字典详情',
    itemDetail: '字典项详情',
    emptyDesc: '请从左侧选择一个分类、字典或字典项查看详情',
    addFirstCategory: '新增第一个分类',
    level: { '1': '分类', '2': '字典', '3': '字典项' },
    field: {
      categoryName: '分类名称',
      categoryCode: '分类代码',
      dictName: '字典名称',
      dictCode: '字典代码',
      itemName: '字典项名称',
      itemCode: '字典项编码',
      category: '所属分类',
      parentDict: '所属字典',
      dictValue: '字典值',
      status: '状态',
      sort: '排序',
      createTime: '创建时间',
      updateTime: '更新时间'
    },
    children: {
      containDict: '包含字典',
      containItem: '包含字典项',
      category: '分类',
      dict: '字典',
      item: '字典项'
    },
    title: '字典管理',
    searchAll: '搜索字典分类、名称、编码、值',
    refreshed: '已刷新',
    confirmDelete: '确定要删除该{level}吗？',
    deleteCancelled: '已取消删除',
    deleteFailed: '删除失败',
    fileName: { category: '字典分类.csv', dict: '字典.csv', item: '字典项.csv' },
    preview: {
      nameSuffix: '名称',
      codeSuffix: '编码',
      more: '还有 {count} 个{child}，点击左侧查看完整列表',
      empty: '该{parent}下暂无{child}'
    },
    treeTitle: '字典结构',
    nodeCount: '{count} 个节点',
    notFound: '未找到匹配的字典',
    validation: {
      nameRequired: '请输入名称',
      nameLength: '名称长度应在 2-20 个字符之间',
      codeRequired: '请输入代码',
      codeLength: '代码长度应在 2-20 个字符之间',
      descriptionMax: '描述长度不能超过 200 个字符',
      statusRequired: '请选择状态',
      valueRequired: '请输入字典值',
      valueLength: '字典值长度应在 1-50 个字符之间',
      sortRequired: '请输入排序'
    }
  },
  log: {
    loginTitle: '登录日志',
    operationTitle: '操作日志',
    loginKeyword: '用户名或IP',
    operationKeyword: '用户名或模块',
    clearLog: '清空日志',
    confirmClearLogin: '确定要清空所有登录日志吗？',
    confirmClearOperation: '确定要清空所有操作日志吗？',
    clearSuccess: '清空成功',
    loginFileName: '登录日志.csv',
    operationFileName: '操作日志.csv',
    dateStart: '开始日期',
    dateEnd: '结束日期',
    dateSeparator: '至',
    field: {
      ip: 'IP地址',
      loginLocation: '登录地点',
      browser: '浏览器',
      os: '操作系统',
      message: '提示信息',
      loginTime: '登录时间',
      operation: '操作内容',
      requestMethod: '请求方式',
      operationLocation: '操作地点',
      duration: '耗时',
      operationTime: '操作时间'
    }
  }
}
