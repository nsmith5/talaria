<template>
  <ul class="nav flex-column">
    <li class="nav-item">
      <a class="nav-link active" href="#">Active</a>
    </li>
    <li class="nav-item">
      <a class="nav-link" href="#">Link</a>
    </li>
    <li class="nav-item">
      <a class="nav-link" href="#">Link</a>
    </li>
    <li class="nav-item">
      <a class="nav-link disabled" href="#">Disabled</a>
    </li>
  </ul>
</template>
<script>
export default {
  name: "menu",
}
</script>
<style>
ul.nav {
  border-right: 1px solid blue;
  height: 100%;
  width: 30%;
}
</style>
