import { AuthClient } from "@/proto/auth_grpc_web_pb";
import { LoginRequest } from "@/proto/auth_pb";

const BACKEND_URL = process.env.API_URL || "http://localhost:8081";
const client = new AuthClient(BACKEND_URL, null, null);

const state = {
  token: "",
  username: "",
  admin: false,
};

const getters = {
  getAuthToken: state => state.token,
  getUsername: state => state.username,
  isAdmin: state => state.admin
};

const actions = {
  login({ commit }, {username, password, $router}) {
    let req = new LoginRequest();
    req.setUsername(username);
    req.setPassword(password);
    client.login(req, {}, (err, resp) => {
      let token = resp.getToken();
      commit("SET_TOKEN", token);
      $router.push('/')
    });
  },
  logout({ commit }, { $router }) {
    commit('SET_TOKEN', '')
    $router.push('/login')
  }
};

const mutations = {
  SET_TOKEN: (state, token) => {
    if (!token) { 
      state.token = ''
      state.username = ''
      state.admin = false
      return
    }
    state.token = token
    let payload = token.split('.')[1]
    let { sub, admin } = JSON.parse(window.atob(payload))
    state.username = sub
    state.admin = admin
  }
};

export default {
  state,
  getters,
  actions,
  mutations
};
