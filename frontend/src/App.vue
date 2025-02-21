<template>
  <div class="host-management">
    <div class="operation-bar">
      <el-button type="primary" @click="handleAdd">添加主机</el-button>
      <el-button type="success" @click="loadHostList">刷新主机</el-button>
    </div>

    <el-table :data="hostList" border style="width: 100%">
      <el-table-column prop="name" label="主机名称" width="180" />
      <el-table-column prop="ip" label="IP地址" width="150" />
      <el-table-column prop="version" label="版本" width="100" />
      <el-table-column prop="cpu" label="CPU" width="100" />
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-button size="small" type="success" v-if="row.status==='online'" @click="handleClose(row)">在线</el-button>
          <el-button size="small" type="danger" v-if="row.status==='offline'" @click="handleStart(row)">离线</el-button>
        </template>
      </el-table-column>
      <el-table-column prop="desc" label="描述" />
      <el-table-column label="操作" width="250">
        <template #default="{ row }">
          <el-button size="small" @click="handleEdit(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row.id)">删除</el-button>
          <el-button size="small" type="primary" @click="handleInstall(row)">配置</el-button>
          <el-button size="small" type="success" @click="handleConnect(row)">连接</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑主机' : '添加主机'" width="600px" top="10vh">
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="80px">
        <el-row :gutter="20">
          <el-col :span="24">
            <el-form-item label="主机名称" prop="name">
              <el-input v-model="formData.name" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="IP地址" prop="ip">
              <el-input v-model="formData.ip" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="用户名" prop="username">
              <el-input v-model="formData.username" />
            </el-form-item>
          </el-col>
          <el-col :span="12" v-if="formData.install == 'wait' || formData.install == 'doing'">
            <el-form-item label="密码" prop="password" >
              <el-input v-model="formData.password" type="password" show-password />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="描述" prop="desc">
              <el-input v-model="formData.desc" />
            </el-form-item>
          </el-col>
          
          <el-col :span="24">
            <el-form-item label="IP" prop="ips">
              <template v-if="typeof formData.ips != 'string'">
              <div style="margin-top: 5px;" v-for="(item,index) in formData.ips"> 
                <el-input v-model="item.ip" style="float: left;width: 160px;margin-right: 5px;" /> 
                <el-input v-model="item.iface" style="float: left;width: 80px;margin-right: 5px;" />
              </div>
              </template>
              <div style="margin-top: 5px;"> 
                <el-input v-model="newip.ip" style="float: left;width: 160px;margin-right: 5px;" /> 
                <el-input v-model="newip.iface" style="float: left;width: 80px;margin-right: 5px;" /> 
                <el-button type="info" style="float: left;width: 60px;" @click="addIP(formData.ip,formData.password||'')">增加</el-button>
              </div>
            </el-form-item>
          </el-col>
        </el-row>
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
import {Install,RunServer,CloseServer,CheckPort,AddServerIP} from '../wailsjs/go/main/App'

interface ipface {
  ip: string
  iface: string
}
interface Host {
  id?: number
  name: string
  ip: string
  port: number
  username: string
  version: string
  password?: string
  desc: string
  ips: ipface[] | ""
  rpc: string
  grpc: string
  cpu: string
  token: string
  status: 'online' | 'offline'
  install: 'doing' | 'finish' | 'wait'
}

const formRef = ref<FormInstance>()
const hostList = ref<Host[]>([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const currentEditId = ref<number>(0)

const newip = ref<ipface>({iface:'',ip:''})

const formData = reactive<Omit<Host, 'id' | 'status'>>({
  name: '',
  ip: '',
  port: 22,
  username: 'root',
  password: '',
  version: "1.0.4",
  desc: '',
  ips: '',
  rpc: '',
  token:'',
  grpc: "",
  cpu: "",
  install: "wait",
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
    host.ips = ""
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
    hosts.forEach(async (host,index) => {
      try {
        const response = await fetch(`http://${host.ip}:5189/metrics`, {
          method: 'GET',
          headers: {'Content-Type': 'application/json'},
        });
        if (response.ok) {
          const content = await response.json();
          console.log(content);
          var ips = content.ips || [];
          host.status = 'online';
          host.cpu =  (content.cpu).toFixed(2) + '%';
          host.ips =  ips;
          host.token =  content.token;
          hostList.value[index] = host;
        } else {
          host.status = 'offline';
          host.cpu = "";
          hostList.value[index] = host;
        }
      } catch (error) {
        host.status = 'offline';
        host.cpu = "";
        hostList.value[index] = host;
      }
    })
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
  if(row.ips.length>0){
    var ip = row.ips[0];
    if(typeof ip === 'string'){
      newip.value.iface = "";
    }else{
      newip.value.iface = ip.iface;
    }
  }else{
    newip.value.iface = "";
  }
  newip.value.ip = "";
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
  const index = hostList.value.findIndex(h => h.id === host.id);
  try {
    CloseServer(host.ip, host.password || '');
    let timerId: number;
    let attemptCount = 0; // 添加一个计数器

    const startTimer = () => {
      timerId = window.setInterval(async () => {
        await fetch(`http://${host.ip}:5189/metrics`, {
          method: 'GET',
          headers: {'Content-Type': 'application/json'},
        }).then(res=>{
          attemptCount++; // 每次查询失败后增加计数器
          if (attemptCount >= 10) { // 如果达到10次查询
            clearInterval(timerId);
            ElMessage.error('关闭失败');
          }
        }).catch(err=>{
          clearInterval(timerId);
          hostList.value[index].status = 'offline';
          ElMessage.success('关闭成功');
        })
      }, 1000); // 每1秒执行一次
    };

    startTimer();
  } catch (error) {
    console.error('关闭失败:', error);
    ElMessage.error('关闭失败');
  }
};

const handleStart = async (host: Host) => {
  ElMessage.info(`正在连接 ${host.ip}:${host.port}`);
  const index = hostList.value.findIndex(h => h.id === host.id);
  try {
    var token = generateRandomString(16);
    RunServer(host.ip, host.password || '', token);
    hostList.value[index].token = token;
    let timerId: number;
    let attemptCount = 0; // 添加一个计数器

    const startTimer = () => {
      timerId = window.setInterval(async () => {
        const res = await CheckPort(host.ip, host.password || '');
        console.log(res);
        if (res === 'success') {
          clearInterval(timerId);
          hostList.value[index].status = 'online';
          ElMessage.success('启动成功');
        } else {
          attemptCount++;
          if (attemptCount >= 10) {
            clearInterval(timerId);
            ElMessage.error('启动失败');
          }
        }
      }, 1000); // 每1秒执行一次
    };

    startTimer();
  } catch (error) {
    console.error('启动失败:', error);
    ElMessage.error('启动失败');
  }
};

const handleInstall = (host: Host) => {
  if(host.status == 'offline'){
    return ElMessage.error('服务器未启动');
  }
  var url = 'http://'+host.ip+':5189?model=&code='+host.token;;
  if (isMacOS() == true) {
    OpenURL(url);
  }else if (isWindows() == true) {
    const windowFeatures = 'width=800,height=600,resizable=yes,scrollbars=yes';
    const newWindow = window.open(url, '', windowFeatures);
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
  var url = 'http://'+host.ip+':5189?model=run&code='+host.token;
  if (isMacOS() == true) {
    OpenURL(url);
  }else if (isWindows() == true) {
    const windowFeatures = 'width=800,height=600,resizable=yes,scrollbars=yes';
    const newWindow = window.open(url, '', windowFeatures);
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

const addIP = async (ip:string, password:string) => {
  try {
    AddServerIP(ip,password,newip.value.ip,newip.value.iface).then((res:any) => {
      console.log(res);
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

const generateRandomString = (length: number) => {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    let randomString = '';

    for (let i = 0; i < length; i++) {
        const randomIndex = Math.floor(Math.random() * chars.length);
        randomString += chars[randomIndex];
    }

    return randomString;
}

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