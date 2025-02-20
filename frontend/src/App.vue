<template>
  <div class="host-management">
    <div class="operation-bar">
      <el-button type="primary" @click="handleAdd">添加主机</el-button>
      <el-button type="primary" @click="handleCheck">检查连接</el-button>
    </div>

    <el-table :data="hostList" border style="width: 100%">
      <el-table-column prop="name" label="主机名称" width="180" />
      <el-table-column prop="ip" label="IP地址" width="150" />
      <el-table-column prop="version" label="版本" width="100" />
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-button size="small" type="success" v-if="row.status==='online'" @click="handleClose(row)">在线</el-button>
          <el-button size="small" type="success" v-if="row.status==='offline'" @click="handleStart(row)">离线</el-button>
        </template>
      </el-table-column>
      <el-table-column prop="desc" label="描述" />
      <el-table-column label="操作" width="250">
        <template #default="{ row }">
          <el-button size="small" @click="handleEdit(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row.id)">删除</el-button>
          <el-button size="small" type="success" @click="handleInstall(row)">配置</el-button>
          <el-button size="small" type="success" @click="handleConnect(row)">连接</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑主机' : '添加主机'" width="600px" top="10vh">
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="80px">
        <el-form-item label="主机名称" prop="name">
          <el-input v-model="formData.name" />
        </el-form-item>
        <el-form-item label="IP地址" prop="ip">
          <el-input v-model="formData.ip" />
        </el-form-item>
        <el-form-item label="端口" prop="port">
          <el-input v-model.number="formData.port" />
        </el-form-item>
        <el-form-item label="用户名" prop="username">
          <el-input v-model="formData.username" />
        </el-form-item>
        <el-form-item label="密码" prop="password" v-if="formData.install == 'wait' || formData.install == 'doing'" >
          <el-input v-model="formData.password" type="password" show-password />
        </el-form-item>
        <el-form-item label="描述" prop="desc">
          <el-input v-model="formData.desc" />
        </el-form-item>
      </el-form>
      <template #footer>
        <div style="display: flex; justify-content: space-between; align-items: center;">
          <el-button type="info" @click="handleTestConnection" v-if="formData.install == 'wait'" >安装插件</el-button>
          <el-button type="info" v-if="formData.install == 'doing'" >安装中</el-button>
          <el-button type="info" v-if="formData.install == 'finish'" >安装完成</el-button>
          <div>
            <el-button @click="dialogVisible = false">取消</el-button>
            <el-button type="primary" @click="submitForm">确认</el-button>
          </div>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { isMacOS,isWindows,OpenURL,loadEnvironment } from './utils/platform';
import { ref, reactive, onMounted } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage, ElMessageBox, install } from 'element-plus'
import {Install,RunServer,CloseServer,CheckPort} from '../wailsjs/go/main/App'

interface Host {
  id?: number
  name: string
  ip: string
  port: number
  username: string
  version: string
  password?: string
  desc: string
  ips: string
  rpc: string
  grpc: string
  status: 'online' | 'offline'
  install: 'doing' | 'finish' | 'wait'
}

const formRef = ref<FormInstance>()
const hostList = ref<Host[]>([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const currentEditId = ref<number>(0)

const formData = reactive<Omit<Host, 'id' | 'status'>>({
  name: '',
  ip: '',
  port: 22,
  username: 'root',
  password: '',
  version:"1.0.4",
  desc: '',
  ips:'',
  rpc: '',
  grpc:"",
  install:"wait",
})

const formRules = reactive<FormRules>({
  name: [{ required: true, message: '请输入主机名称', trigger: 'blur' }],
  ip: [
    { required: true, message: '请输入IP地址', trigger: 'blur' },
    { pattern: /^(\d{1,3}\.){3}\d{1,3}$/, message: '请输入正确的IP地址', trigger: 'blur' }
  ],
  port: [
    { required: true, message: '请输入端口号', trigger: 'blur' },
    { type: 'number', message: '端口号必须为数字', trigger: 'blur' }
  ],
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }]
})

const initDB = (): Promise<IDBDatabase> => {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open('HostDB', 1)

    request.onupgradeneeded = (event) => {
      const db = (event.target as IDBOpenDBRequest).result
      if (!db.objectStoreNames.contains('hosts')) {
        const store = db.createObjectStore('hosts', { keyPath: 'id', autoIncrement: true })
        store.createIndex('ip', 'ip', { unique: false })
      }
    }

    request.onsuccess = () => resolve(request.result)
    request.onerror = () => reject(request.error)
  })
}

const getAllHosts = async (): Promise<Host[]> => {
  const db = await initDB()
  return new Promise((resolve, reject) => {
    const transaction = db.transaction('hosts', 'readonly')
    const store = transaction.objectStore('hosts')
    const request = store.getAll()

    request.onsuccess = () => resolve(request.result)
    request.onerror = () => reject(request.error)
  })
}

const saveHost = async (host: Host): Promise<void> => {
  const db = await initDB()
  return new Promise((resolve, reject) => {
    const transaction = db.transaction('hosts', 'readwrite')
    const store = transaction.objectStore('hosts')
    const request = host.id ? store.put(host) : store.add(host)

    request.onsuccess = () => resolve()
    request.onerror = () => reject(request.error)
  })
}

const deleteHost = async (id: number): Promise<void> => {
  const db = await initDB()
  return new Promise((resolve, reject) => {
    const transaction = db.transaction('hosts', 'readwrite')
    const store = transaction.objectStore('hosts')
    const request = store.delete(id)

    request.onsuccess = () => resolve()
    request.onerror = () => reject(request.error)
  })
}

const loadHostList = async () => {
  try {
    const hosts = await getAllHosts();
    // 确保每个主机的 status 字段都有一个默认值 'offline'
    hostList.value = hosts.map(host => ({
      ...host,
      status: 'offline', // 如果 status 不存在，则设置为 'offline'
      install: host.install, // 如果 install 不存在，则设置为 'wait'
    }));
  } catch (error) {
    console.error('加载数据失败:', error);
    ElMessage.error('数据加载失败');
  }
};

const getHostStatus = async (host:string) => {
  try {
    const response = await fetch("http://"+host+":5189/metrics", {
      method: 'GET', // 指定请求方法为 POST
      headers: {
        'Content-Type': 'application/json', // 设置请求头，指定发送的数据格式为 JSON
      },
    });
    if (response.ok) {
      const content = await response.json();
      console.log(content)
    } else {
      console.error('Failed to save config:', response.statusText);
    }
  } catch (error) {
    console.error('Error saving config:', error);
  }
  return 'offline';
} 
const handleCheck = async () => {
  try {
    const hosts = await getAllHosts();

    // 使用 Promise.all 来并发请求每个主机的状态
    const updatedHosts = await Promise.all(
      hosts.map(async (host) => {
        try {
          // 假设 getHostStatus 是一个异步函数，用于获取主机的状态
          const status = await getHostStatus(host.ip);
          return {
            ...host,
            status: 'offline', // 如果接口返回的状态为空，则设置为 'offline'
            install: 'finish', // 如果 install 不存在，则设置为 'wait'
          };
        } catch (error) {
          console.error(`获取主机 ${host.id} 状态失败:`, error);
          return {
            ...host,
            status: 'offline', // 如果请求失败，则设置为 'offline'
            install: 'finish', // 如果 install 不存在，则设置为 'wait'
          };
        }
      })
    );

    //hostList.value = updatedHosts;
  } catch (error) {
    console.error('加载数据失败:', error);
    ElMessage.error('数据加载失败');
  }
};
const handleAdd = () => {
  isEdit.value = false
  dialogVisible.value = true
  Object.assign(formData, {
    name: '127.0.0.1',
    ip: '127.0.0.1',
    port: 22,
    username: 'root',
    password: 'root',
    desc: '',
    install:"wait"
  })
}

const handleEdit = (row: Host) => {
  isEdit.value = true
  currentEditId.value = row.id!
  Object.assign(formData, row)
  dialogVisible.value = true
}

const handleDelete = async (id: number) => {
  ElMessageBox.confirm('确认删除该主机？', '警告', {
    confirmButtonText: '确认',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      await deleteHost(id)
      await loadHostList()
      ElMessage.success('删除成功')
    } catch (error) {
      ElMessage.error('删除失败')
    }
  })
}
const handleClose = async (host: Host) => {
  ElMessage.info(`正在关闭 ${host.ip}:${host.port}`);

  try {
    CloseServer(host.ip, host.password || '');
    let timerId: number;
    let timeoutId: number;
    const startTimer = () => {
      timerId = window.setInterval(async () => {
        const res = await CheckPort(host.ip, host.password || '');
        console.log('CloseServer CheckPort',res);
        if (res === 'false') {
          clearInterval(timerId);
          clearTimeout(timeoutId);
          // 更新主机状态为 'offline'
          const index = hostList.value.findIndex(h => h.id === host.id);
          if (index !== -1) {
            hostList.value[index].status = 'offline';
          }
          ElMessage.success('关闭成功');
        }
      }, 1000); // 每1秒执行一次
    };

    const startTimeout = () => {
      timeoutId = window.setTimeout(() => {
        clearInterval(timerId);
        ElMessage.warning('关闭超时，正在重新关闭...');
        handleClose(host); // 重新关闭
      }, 10000); // 10秒超时
    };

    startTimer();
    startTimeout();
  } catch (error) {
    console.error('关闭失败:', error);
    ElMessage.error('关闭失败');
  }
}
const handleStart = async (host: Host) => {
  ElMessage.info(`正在连接 ${host.ip}:${host.port}`);

  try {
    RunServer(host.ip, host.password || '');

    let timerId: number;
    let timeoutId: number;

    const startTimer = () => {
      timerId = window.setInterval(async () => {
        const res = await CheckPort(host.ip, host.password || '');
        console.log(res);
        if (res === 'success') {
          clearInterval(timerId);
          clearTimeout(timeoutId);
          // 更新主机状态为 'online'
          const index = hostList.value.findIndex(h => h.id === host.id);
          if (index !== -1) {
            hostList.value[index].status = 'online';
          }
          ElMessage.success('启动成功');
        }
      }, 1000); // 每1秒执行一次
    };

    const startTimeout = () => {
      timeoutId = window.setTimeout(() => {
        clearInterval(timerId);
        ElMessage.warning('启动超时，正在重新连接...');
        handleStart(host); // 重新连接
      }, 10000); // 10秒超时
    };

    startTimer();
    startTimeout();
  } catch (error) {
    console.error('启动失败:', error);
    ElMessage.error('启动失败');
  }
};

const handleInstall = (host: Host) => {
  if(host.status == 'offline'){
    return ElMessage.error('服务器未启动');
  }

  if (isMacOS() == true) {
    OpenURL('http://'+host.ip+':5189/install?model=');
  }else if (isWindows() == true) {
    const windowFeatures = 'width=800,height=600,resizable=yes,scrollbars=yes';
    const newWindow = window.open('http://'+host.ip+':5189/install?model=', '', windowFeatures);

    if (newWindow) {
      // 设置新窗口的标题
      newWindow.document.title = `连接主机 - ${host.name}`;
      
      // 添加加载中的提示
      newWindow.document.body.innerHTML = `
        <div style="display: flex; justify-content: center; align-items: center; height: 100vh;">
          <h2>正在连接到 ${host.ip}:${host.port}...</h2>
        </div>
      `;
    } else {
      ElMessage.error('无法打开新窗口，请检查浏览器设置');
    }
  }
}
const handleConnect = (host: Host) => {
  if(host.status == 'offline'){
    return ElMessage.error('服务器未启动');
  }
  if (isMacOS() == true) {
    OpenURL('http://'+host.ip+':5189/install?model=');
  }else if (isWindows() == true) {
    // 打开新窗口
    const windowFeatures = 'width=800,height=600,resizable=yes,scrollbars=yes';
    const newWindow = window.open('http://'+host.ip+':5189?model=run', '', windowFeatures);

    if (newWindow) {
      // 设置新窗口的标题
      newWindow.document.title = `连接主机 - ${host.name}`;
      
      // 添加加载中的提示
      newWindow.document.body.innerHTML = `
        <div style="display: flex; justify-content: center; align-items: center; height: 100vh;">
          <h2>正在连接到 ${host.ip}:${host.port}...</h2>
        </div>
      `;

      // 预留后续逻辑
      // 这里可以添加 WebSocket 连接、终端初始化等逻辑
      // 例如：
      // initTerminal(newWindow, host);
    } else {
      ElMessage.error('无法打开新窗口，请检查浏览器设置');
    }
  }
}

const handleTestConnection = async () => {
  if (!formRef.value) return;
  try {
    // 修复：如果是新增操作，移除 id 字段
    const hostData: Host = {
      ...formData,
      status: 'offline'
    };
    // 使用 Vue 的响应式方法确保更新
    formData.install = "doing";

    if(hostData.password==''){
      formData.install = "wait";
      return ElMessage.error('密码不能为空');
    }
    ElMessage.info(`正在连接 ${hostData.ip}`);
    Install(hostData.ip, hostData.password || '').then((res) => {
      if (res === 'success') {
        // 使用 Vue 的响应式方法确保更新
        formData.install = "finish";
        ElMessage.success('安装成功');
      } else {
        // 使用 Vue 的响应式方法确保更新
        formData.install = "wait";
        ElMessage.error('安装失败' + res);
      }
    });
  } catch (error) {
    console.error('操作失败:', error);
    ElMessage.error('操作失败');
  }
};

const submitForm = async () => {
  if (!formRef.value) return;

  try {
    await formRef.value.validate();

    // 修复：如果是新增操作，移除 id 字段
    const hostData: Host = {
      ...formData,
      status: 'online',
      install: "finish"
    };

    if (isEdit.value && currentEditId.value) {
      // 编辑操作时，保留 id
      hostData.id = currentEditId.value;
    }

    await saveHost(hostData);
    await loadHostList();

    dialogVisible.value = false;
    ElMessage.success(isEdit.value ? '修改成功' : '添加成功');
  } catch (error) {
    console.error('操作失败:', error);
    ElMessage.error('操作失败');
  }
};

onMounted(async() => {
  await loadEnvironment()
  loadHostList()
})
</script>

<style scoped>
.host-management {
  padding: 20px;
}

.operation-bar {
  margin-bottom: 20px;
  float: left;
}

.el-table {
  margin-top: 20px;
}

.el-button + .el-button {
  margin-left: 8px;
}
</style>