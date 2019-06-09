<?php

use Illuminate\Http\Request;
use Psy\Util\Json;

/*
|--------------------------------------------------------------------------
| API Routes
|--------------------------------------------------------------------------
|
| Here is where you can register API routes for your application. These
| routes are loaded by the RouteServiceProvider within a group which
| is assigned the "api" middleware group. Enjoy building your API!
|
*/
Route::middleware('auth:api')->get('/user', function (Request $request) {
    return $request->user();
});

Route::group(['namespace' => 'back_api'], function () {
    Route::post('/login', 'AdminController@login');
});

Route::group(['prefix' => 'cart', 'middleware' => ['client.credentials']], function () {
    return false;
});