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
      <el-table-column prop="pid" label="进程" width="100">
        <template #default="{ row }">
          <el-tag v-if="row.pid==-2">无</el-tag>
          <el-tag v-else-if="row.pid==0">未启动</el-tag>
          <el-tag v-else-if="row.pid==-1">未安装</el-tag>
          <el-tag v-else>{{row.pid}}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="130">
        <template #default="{ row }">
          <el-button size="small" type="success" v-if="row.status==='online'" @click="handleClose(row)">在线</el-button>
          <el-button size="small" type="danger" v-if="row.status==='offline'" @click="handleStart(row)">离线</el-button>
          <el-button size="small" type="success" v-if="row.status==='onling'">开启中</el-button>
          <el-button size="small" type="danger" v-if="row.status==='offling'">关闭中</el-button>
          <el-button size="small" type="danger" v-if="row.status==='offline' && row.pid>0" @click="handleClose(row)">关闭</el-button>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200">
        <template #default="{ row }">
          <el-button size="small" @click="handleEdit(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row.id)">删除</el-button>
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
          <el-col  :span="24">
            <el-form-item label="描述" prop="desc">
              <el-input v-model="formData.desc" />
            </el-form-item>
          </el-col>
          <el-col  :span="24" v-if="formData.token">
            <el-form-item label="安全码" prop="token">
              <el-input v-model="formData.token" />
            </el-form-item>
          </el-col>
          <el-col  :span="24">
            <el-form-item label="私钥" prop="siyao">
              <div style="margin-top: 5px;"> 
                <el-input v-model="siyao" type="password" style="float: left;width: 180px;margin-right: 5px;" /> 
                <el-button type="info" v-if="dosiyao==false" style="float: left;width: 60px;" @click="setSiyao(formData.ip,formData.password||'')">更新</el-button>
                <el-button type="info" v-if="dosiyao==true" style="float: left;width: 60px;">更新中</el-button>
              </div>
            </el-form-item>
          </el-col>
          
          <el-col :span="24">
            <el-form-item label="日志" prop="log">
              <el-input type="textarea" row="4" v-model="formData.log" />
            </el-form-item>
          </el-col>
          
          <el-col :span="24" v-if="formData.token">
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
          <el-button type="info" v-if="formData.install == 'uninstall'" >卸载中</el-button>
          <el-button type="info" @click="handleUninstall"  v-if="formData.install == 'finish'" >卸载</el-button>
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
import {Install,Uninstall,RunServer,CloseServer,CheckPort,AddServerIP,Setprivatekey} from '../wailsjs/go/main/App'

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
  pid:number,
  grpc: string
  cpu: string
  token: string
  status: 'online' | 'offline' | 'offing' | 'oning'
  install: 'doing' | 'finish' | 'wait' | 'uninstall'
  log:string
}

const formRef = ref<FormInstance>()
const hostList = ref<Host[]>([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const siyao = ref("")
const dosiyao = ref(false)
const currentEditId = ref<number>(0)
  import axios from 'axios';
const newip = ref<ipface>({iface:'',ip:''})

const formData = reactive<Omit<Host, 'id' | 'status'>>({
  name: '',
  ip: '',
  port: 22,
  username: 'root',
  password: '',
  version: "1.0.5",
  desc: '',
  ips: '',
  rpc: '',
  token:'',
  grpc: "",
  cpu: "",
  pid:-2,
  install: "wait",
  log:''
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
    const updatedHosts = await Promise.all(hosts.map(async (host) => {
      try {
        const response = await axios.get(`http://${host.ip}:5189/metrics`, {
          headers: {'Content-Type': 'application/json'},
        });
        const content = response.data; // axios 自动解析 JSON 数据
        host.status = 'online';
        host.token = content.token;
        host.cpu = (content.cpu).toFixed(2) + '%';
        host.ips = content.ips || [];
        console.log(host)
        const response2 = await axios.get(`http://${host.ip}:5189/pid`, {
          headers: {'Content-Type': 'application/json'},
        });
        host.pid = Number(response2.data.pid) || -2;
        console.log(host)
      } catch (err) {
        host.status = 'offline';
      }
      return host;
    }));
    hostList.value = updatedHosts;
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
      const index = hostList.value.findIndex(host => host.id === id);
      hostList.value.splice(index, 1);
      await loadHostList()
      ElMessage.success('删除成功')
    } catch (error) {
      ElMessage.error('删除失败')
    }
  })
}

// const handleUp = async (host: Host) => {
//   const index = hostList.value.findIndex(h => h.id === host.id);
//   hostList.value[index].status = 'online';
// }


const handleClose = async (host: Host) => {
  ElMessage.info(`正在关闭 ${host.ip}:${host.port}`);

  // 找到当前主机在列表中的索引
  const index = hostList.value.findIndex(h => h.id === host.id);
  if (index === -1) return; // 如果找不到主机，直接返回

  // 更新主机状态为 "关闭中"
  hostList.value[index].status = 'offing';

  try {
    // 调用关闭服务器的函数
    CloseServer(host.ip, host.password || '');

    let attemptCount = 0; // 尝试次数计数器
    const maxAttempts = 10; // 最大尝试次数
    const pollInterval = 1000; // 轮询间隔时间（1秒）

    // 定义轮询函数
    const pollServerStatus = async () => {
      try {
        // 发送请求检查服务器状态
        const response = await axios.get(`http://${host.ip}:5189/metrics`, {
          headers: {'Content-Type': 'application/json'},
        });

        // 如果服务器仍在运行，增加尝试次数
        attemptCount++;
        if (attemptCount >= maxAttempts) {
          // 达到最大尝试次数，停止轮询并提示关闭失败
          ElMessage.error('关闭失败');
          hostList.value[index].status = 'offline'; // 更新状态为离线
          return;
        }

        // 继续轮询
        setTimeout(pollServerStatus, pollInterval);
      } catch (err) {
        // 如果请求失败，说明服务器已关闭
        ElMessage.success('关闭成功');
        hostList.value[index].status = 'offline'; // 更新状态为离线
      }
    };

    // 开始轮询
    pollServerStatus();
  } catch (error) {
    // 处理关闭过程中的异常
    console.error('关闭失败:', error);
    ElMessage.error('关闭异常');
    hostList.value[index].status = 'offline'; // 更新状态为离线
  }
};

const handleStart = async (host: Host) => {
  ElMessage.info(`正在连接 ${host.ip}:${host.port}`);

  // 找到当前主机在列表中的索引
  const index = hostList.value.findIndex(h => h.id === host.id);
  if (index === -1) return; // 如果找不到主机，直接返回

  // 更新主机状态为 "启动中"
  hostList.value[index].status = 'oning';

  try {
    // 生成随机 token
    const token = generateRandomString(16);
    hostList.value[index].token = token;

    // 启动服务器
    const res = await RunServer(host.ip, host.password || '', token);
    console.log(res);
    hostList.value[index].log = res;

    let attemptCount = 0; // 尝试次数计数器
    const maxAttempts = 10; // 最大尝试次数
    const pollInterval = 1000; // 轮询间隔时间（1秒）

    // 定义轮询函数
    const pollServerStatus = async () => {
      try {
        // 检查端口状态
        const res = await CheckPort(host.ip, host.password || '');
        console.log(res);

        if (res === 'success') {
          // 如果启动成功，停止轮询并更新状态
          ElMessage.success('启动成功');
          hostList.value[index].status = 'online';
          return;
        } else {
          // 如果未成功，增加尝试次数
          attemptCount++;
          if (attemptCount >= maxAttempts) {
            // 达到最大尝试次数，停止轮询并提示启动失败
            ElMessage.error('启动失败');
            hostList.value[index].status = 'offline'; // 更新状态为离线
            return;
          }

          // 继续轮询
          setTimeout(pollServerStatus, pollInterval);
        }
      } catch (err) {
        // 如果检查端口状态失败，停止轮询并提示启动异常
        console.error('检查端口状态失败:', err);
        ElMessage.error('启动异常');
        hostList.value[index].status = 'offline'; // 更新状态为离线
      }
    };

    // 开始轮询
    pollServerStatus();
  } catch (error) {
    // 处理启动过程中的异常
    console.error('启动失败:', error);
    ElMessage.error('启动异常');
    hostList.value[index].status = 'offline'; // 更新状态为离线
  }
};

const handleConnect = (host: Host) => {
  const index = hostList.value.findIndex(h => h.id === host.id);
  var url = 'http://'+host.ip+':5189?model=run&code='+host.token;
  axios.get("http://"+host.ip+":5189/metrics", {
    headers: {'Content-Type': 'application/json'}
  })
  .then(() => {
    host.status = 'online';
    hostList.value[index] = host;
    if (isMacOS() == true) {
      OpenURL(url);
    }else if (isWindows() == true) {
      const windowFeatures = 'width=900,height=600,resizable=yes,scrollbars=yes';
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
  }).catch(async()=>{
    OpenURL(url);
  })
}

const handleUninstall = async () => {
  if (!formRef.value) return;
  try {
    // 修复：如果是新增操作，移除 id 字段
    const hostData: Host = {
      ...formData,
      status: 'offline'
    };
    formData.install = "uninstall";
    if(hostData.password==''){
      return ElMessage.error('密码不能为空');
    }
    ElMessage.info(`正在卸载 ${hostData.ip}`);
    Uninstall(hostData.ip, hostData.password || '').then((res) => {
      console.log('Uninstall',res)
      if (res === 'success') {
        formData.install = "wait";
        ElMessage.success('安装成功');
      } else {
        formData.install = "finish";
        ElMessage.error('安装失败' + res);
      }
    });
  } catch (error) {
    console.error('操作失败:', error);
    ElMessage.error('操作失败');
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
const setSiyao = async (ip:string, password:string) => {
  try {
    dosiyao.value = true
    Setprivatekey(ip,password,siyao.value).then((res:any) => {
      dosiyao.value = false
      if (res === 'success') {
        siyao.value = ""
        ElMessage.success('设置成功');
      } else {
        ElMessage.error('设置失败' + res);
      }
    });
  } catch (error) {
    dosiyao.value = false
    console.error('操作失败:', error);
    ElMessage.error('操作失败');
  }
}

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