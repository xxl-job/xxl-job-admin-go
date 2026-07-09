export default {
  layout: {
    settings: 'Layout Settings',
    showTagsView: 'Show TagsView',
    showBreadcrumb: 'Show Breadcrumb',
    showLogo: 'Show Logo',
    showFooter: 'Show Footer',
    theme: 'Theme',
    primaryColor: 'Primary Color',
    componentSize: 'Component Size',
    locale: 'Language'
  },
  size: {
    large: 'Large',
    default: 'Default',
    small: 'Small'
  },
  common: {
    welcome: 'Welcome, {name}',
    action: {
      search: 'Search',
      reset: 'Reset',
      create: 'Create',
      edit: 'Edit',
      view: 'View',
      delete: 'Delete',
      batchDelete: 'Batch Delete',
      export: 'Export',
      refresh: 'Refresh',
      confirm: 'Confirm',
      cancel: 'Cancel',
      close: 'Close',
      save: 'Save',
      submit: 'Submit'
    },
    status: {
      enable: 'Enabled',
      disable: 'Disabled',
      yes: 'Yes',
      no: 'No',
      success: 'Success',
      failed: 'Failed',
      error: 'Error',
      pending: 'Pending',
      processing: 'Processing',
      completed: 'Completed',
      cancelled: 'Cancelled',
      deleted: 'Deleted'
    },
    table: {
      columnSettings: 'Columns',
      visibleFields: 'Visible Fields',
      showAll: 'Show All',
      noData: 'No Data'
    },
    placeholder: {
      input: 'Please enter {label}',
      select: 'Please select {label}',
      selectRole: 'Select role',
      selectUser: 'Search by name / username',
      selectDept: 'Select department'
    },
    message: {
      createSuccess: 'Created successfully',
      updateSuccess: 'Updated successfully',
      deleteSuccess: 'Deleted successfully',
      deleteCancelled: 'Deletion cancelled',
      createFailed: 'Create failed',
      updateFailed: 'Update failed',
      deleteFailed: 'Delete failed',
      exportSuccess: 'Exported successfully',
      exportFailed: 'Export failed',
      fieldRequired: 'This field is required'
    },
    column: {
      status: 'Status',
      description: 'Description',
      createTime: 'Created',
      updateTime: 'Updated',
      lastLoginTime: 'Last Login',
      actions: 'Actions',
      sort: 'Sort',
      leader: 'Leader',
      publisher: 'Publisher',
      publishTime: 'Publish Time'
    },
    notFound: {
      subTitle: 'Sorry, the page you visited does not exist',
      backHome: 'Back Home'
    }
  },
  auth: {
    title: 'Sign in',
    username: 'Username',
    password: 'Password',
    submit: 'Sign in',
    loginFailed: 'Login failed, please try again later',
    usernameLength: 'Username must be 3-20 characters',
    passwordLength: 'Password must be at least 6 characters'
  },
  user: {
    title: 'User Management',
    searchKeyword: 'Username, name, email or phone',
    addTitle: 'Create User',
    editTitle: 'Edit User',
    viewTitle: 'View User',
    field: {
      username: 'Username',
      realName: 'Name',
      email: 'Email',
      phone: 'Phone',
      roles: 'Roles',
      status: 'Status',
      password: 'Password',
      confirmPassword: 'Confirm Password'
    },
    placeholder: {
      selectRole: 'Select role',
      passwordAdd: 'Please enter password',
      passwordEdit: 'Leave blank to keep unchanged'
    },
    validation: {
      usernameRequired: 'Please enter username',
      usernameLength: 'Username must be 3-20 characters',
      realNameRequired: 'Please enter name',
      realNameLength: 'Name must be 2-20 characters',
      emailRequired: 'Please enter email',
      emailInvalid: 'Please enter a valid email',
      phoneRequired: 'Please enter phone',
      phoneInvalid: 'Please enter a valid phone number',
      rolesRequired: 'Please select role',
      rolesAtLeastOne: 'Please select at least one role',
      statusRequired: 'Please select status',
      passwordRequired: 'Please enter password',
      passwordLength: 'Password must be at least 6 characters',
      confirmPasswordRequired: 'Please enter password again',
      confirmMismatch: 'Passwords do not match'
    }
  },
  role: {
    title: 'Role Management',
    searchKeyword: 'Role name, code or description',
    permissionLabel: 'Permissions',
    permissionConfigTitle: 'Permission Configuration',
    permissionSaveSuccess: 'Permissions saved successfully',
    addTitle: 'Create Role',
    editTitle: 'Edit Role',
    viewTitle: 'View Role',
    field: {
      name: 'Role Name',
      code: 'Role Code',
      description: 'Description',
      status: 'Status'
    },
    validation: {
      nameRequired: 'Please enter role name',
      nameLength: 'Role name must be 2-20 characters',
      codeRequired: 'Please enter role code',
      codeLength: 'Role code must be 2-20 characters',
      descriptionMax: 'Description cannot exceed 200 characters',
      statusRequired: 'Please select status'
    }
  },
  dept: {
    title: 'Department Management',
    searchKeyword: 'Department name',
    addSubLabel: 'Add Sub-department',
    deleteConfirm: 'Delete this department? Sub-departments will also be deleted.',
    addTitle: 'Create Department',
    editTitle: 'Edit Department',
    field: {
      parentId: 'Parent Department',
      name: 'Department Name',
      leader: 'Leader',
      phone: 'Phone',
      email: 'Email',
      sort: 'Sort',
      status: 'Status'
    },
    placeholder: {
      noParent: 'No parent department'
    },
    validation: {
      nameRequired: 'Please enter department name'
    }
  },
  notice: {
    title: 'Notice Management',
    searchKeyword: 'Notice title',
    priority: { high: 'Urgent', medium: 'Important', low: 'Normal' },
    bannerTitle: '{priority} Announcement: {title}',
    actionPublish: 'Publish',
    actionRevoke: 'Revoke',
    center: {
      title: 'Notification Center',
      markAllRead: 'Mark all read',
      viewAll: 'View All',
      publisher: 'Publisher',
      publishTime: 'Publish Time',
      emptyAnnouncement: 'No announcements',
      emptyNotice: 'No notices',
      emptyTodo: 'No todos'
    },
    confirmPublish: 'Publish this notice?',
    confirmRevoke: 'Revoke this notice?',
    publishSuccess: 'Published successfully',
    revokeSuccess: 'Revoked successfully',
    addTitle: 'Create Notice',
    editTitle: 'Edit Notice',
    viewTitle: 'View Notice',
    field: {
      title: 'Title',
      type: 'Type',
      status: 'Status',
      priority: 'Priority',
      content: 'Content',
      publishTime: 'Publish Time',
      expireTime: 'Expire Time'
    },
    option: {
      typeAnnouncement: 'Announcement',
      typeNotice: 'Notice',
      typeTodo: 'Todo',
      statusDraft: 'Draft',
      statusPublished: 'Published',
      statusExpired: 'Expired',
      statusRevoked: 'Revoked',
      priorityHigh: 'High',
      priorityMedium: 'Medium',
      priorityLow: 'Low'
    },
    validation: {
      titleRequired: 'Please enter title',
      contentRequired: 'Please enter content'
    }
  },
  permission: {
    title: 'Permission Management',
    searchKeyword: 'Permission name, code or description',
    addTitle: 'Create Permission',
    editTitle: 'Edit Permission',
    viewTitle: 'View Permission',
    field: {
      name: 'Permission Name',
      code: 'Permission Code',
      module: 'Module',
      status: 'Status',
      description: 'Description'
    },
    option: {
      moduleSystem: 'System',
      moduleUser: 'User',
      moduleRole: 'Role',
      modulePermission: 'Permission',
      moduleDict: 'Dictionary',
      moduleConfig: 'Configuration'
    },
    validation: {
      nameRequired: 'Please enter permission name',
      codeRequired: 'Please enter permission code',
      moduleRequired: 'Please select module',
      statusRequired: 'Please select status',
      descriptionMax: 'Description cannot exceed 200 characters'
    }
  },
  menu: {
    title: 'Menu Management',
    addTitle: 'Create Top Menu',
    addSubTitle: 'Create Submenu - {name}',
    editTitle: 'Edit Menu',
    viewTitle: 'View Menu',
    field: {
      parentId: 'Parent Menu',
      name: 'Menu Name',
      path: 'Route Path',
      component: 'Component Path',
      icon: 'Icon',
      sort: 'Sort',
      status: 'Status'
    },
    placeholder: {
      noParent: 'Leave blank for top-level menu',
      component: 'e.g. dashboard/views/Home',
      icon: 'Element Plus icon name'
    },
    validation: {
      nameRequired: 'Please enter menu name',
      pathRequired: 'Please enter route path',
      statusRequired: 'Please select status'
    },
    actionAddSub: 'Add Submenu',
    loadFailed: 'Failed to load menu tree',
    deleteConfirm: 'Delete menu "{name}"? Submenus will also be deleted.',
    sortUpdated: 'Sort order updated'
  },
  crud: {
    title: 'CRUD Demo',
    addTitle: 'Create Record',
    editTitle: 'Edit Record',
    viewTitle: 'View Record',
    field: {
      name: 'Name',
      province: 'Province',
      city: 'City',
      date: 'Date',
      address: 'Address',
      zip: 'Zip'
    },
    validation: {
      nameRequired: 'Please enter name',
      provinceRequired: 'Please select province',
      cityRequired: 'Please enter city',
      addressRequired: 'Please enter address',
      zipRequired: 'Please enter zip',
      zipPattern: 'Please enter a 6-digit zip code'
    }
  },
  dashboard: {
    defaultUser: 'User',
    roleSuperAdmin: 'Super Admin',
    roleNormal: 'User',
    stat: {
      users: 'Total Users',
      orders: 'Orders',
      revenue: 'Revenue',
      active: 'Active Users'
    },
    quickAction: {
      user: 'User Management',
      role: 'Role Management',
      permission: 'Permission Management',
      dict: 'Dictionary',
      menu: 'Menu Management',
      notice: 'Notice Management',
      crud: 'CRUD Demo',
      profile: 'Profile',
      docs: 'Docs'
    },
    message: {
      inDevelopment: 'This feature is under development',
      docsBuilding: 'Docs site is under construction',
      loadFailed: 'Failed to load dashboard data'
    },
    trendLabel: 'vs yesterday',
    quickEntry: 'Quick Actions',
    noQuickAction: 'No quick actions available',
    latestActivity: 'Latest Activity',
    noActivity: 'No activity',
    visitTrend: 'Visit Trend',
    roleDistribution: 'Role Distribution',
    lastLoginPrefix: 'Last login'
  },
  profile: {
    title: 'Profile',
    editProfile: 'Edit Profile',
    basicInfo: 'Basic Info',
    personalSettings: 'Settings',
    field: {
      username: 'Username',
      email: 'Email',
      phone: 'Phone',
      gender: 'Gender',
      department: 'Department',
      position: 'Position',
      joinDate: 'Join Date',
      lastLogin: 'Last Login',
      name: 'Name',
      language: 'Language',
      timezone: 'Timezone',
      notification: 'Notifications',
      emailReminder: 'Email Reminders'
    },
    option: {
      chinese: 'Chinese',
      english: 'English'
    },
    saveSuccess: 'Profile saved successfully'
  },
  dict: {
    selectNode: 'Select a node',
    categoryDetail: 'Category Details',
    dictDetail: 'Dictionary Details',
    itemDetail: 'Item Details',
    emptyDesc: 'Select a category, dictionary or item on the left to view details',
    addFirstCategory: 'Create first category',
    level: { '1': 'Category', '2': 'Dictionary', '3': 'Item' },
    field: {
      categoryName: 'Category Name',
      categoryCode: 'Category Code',
      dictName: 'Dictionary Name',
      dictCode: 'Dictionary Code',
      itemName: 'Item Name',
      itemCode: 'Item Code',
      category: 'Category',
      parentDict: 'Dictionary',
      dictValue: 'Value',
      status: 'Status',
      sort: 'Sort',
      createTime: 'Created',
      updateTime: 'Updated'
    },
    children: {
      containDict: 'Dictionaries',
      containItem: 'Items',
      category: 'Category',
      dict: 'Dictionary',
      item: 'Item'
    },
    title: 'Dictionary Management',
    searchAll: 'Search category, name, code, value',
    refreshed: 'Refreshed',
    confirmDelete: 'Delete this {level}?',
    deleteCancelled: 'Deletion cancelled',
    deleteFailed: 'Delete failed',
    fileName: { category: 'dictionary-category.csv', dict: 'dictionary.csv', item: 'dictionary-item.csv' },
    preview: {
      nameSuffix: ' Name',
      codeSuffix: ' Code',
      more: '{count} more {child}, click left for full list',
      empty: 'No {child} under this {parent}'
    },
    treeTitle: 'Dictionary Tree',
    nodeCount: '{count} nodes',
    notFound: 'No matching dictionary found',
    validation: {
      nameRequired: 'Please enter name',
      nameLength: 'Name must be 2-20 characters',
      codeRequired: 'Please enter code',
      codeLength: 'Code must be 2-20 characters',
      descriptionMax: 'Description cannot exceed 200 characters',
      statusRequired: 'Please select status',
      valueRequired: 'Please enter value',
      valueLength: 'Value must be 1-50 characters',
      sortRequired: 'Please enter sort'
    }
  },
  log: {
    loginTitle: 'Login Log',
    operationTitle: 'Operation Log',
    loginKeyword: 'Username or IP',
    operationKeyword: 'Username or module',
    clearLog: 'Clear Log',
    confirmClearLogin: 'Clear all login logs?',
    confirmClearOperation: 'Clear all operation logs?',
    clearSuccess: 'Cleared successfully',
    loginFileName: 'login-log.csv',
    operationFileName: 'operation-log.csv',
    dateStart: 'Start date',
    dateEnd: 'End date',
    dateSeparator: 'to',
    field: {
      ip: 'IP Address',
      loginLocation: 'Location',
      browser: 'Browser',
      os: 'OS',
      message: 'Message',
      loginTime: 'Login Time',
      operation: 'Operation',
      requestMethod: 'Method',
      operationLocation: 'Location',
      duration: 'Duration',
      operationTime: 'Operation Time'
    }
  }
}
