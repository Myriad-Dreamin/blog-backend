/*jshint esversion: 6 */

import VueRouter from 'vue-router';

import ExampleComponent from './components/ExampleComponent.vue';

let router = [
    {
        path : '/',
        name: 'home',
        component: ExampleComponent
    },
    {
        path : '/articles',
        name: 'ff',
        component: ExampleComponent
    }
];

export default new VueRouter({
    router
});

