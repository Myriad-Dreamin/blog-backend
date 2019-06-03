<?php

/*
|--------------------------------------------------------------------------
| Web Routes
|--------------------------------------------------------------------------
|
| Here is where you can register web routes for your application. These
| routes are loaded by the RouteServiceProvider within a group which
| contains the "web" middleware group. Now create something great!
|
*/

Route::get('/', function () {
    return view('welcome');
});

Route::get('/about', function () {
    // with -> (key string, value)
    // with -> (key string, [
    // ke1 => va1
    //])
    
    // in app.blade.php
    // @yield('content')
    

    // in some.blade.php
    //
    // @extends('app')
    // @section('content')
    // @stop

    // @if(php-expression)
    // @else
    // @endif

    // @foreach( v as v-arr)
    // @endforeach
    
    // env(key value)

    // compact value
    return view('sites.about');
});


