<?php

namespace App\Http\Controllers\back_api;

use Illuminate\Http\Request;
use App\Http\Controllers\Controller;

class AdminController extends Controller
{
    protected $proxy;

    public function __construct()
    {
        $this->proxy = new \App\Http\Proxy\AdminProxy(new \GuzzleHttp\Client());
    }

    public function login(Request $request)
    {
        return $this->proxy->proxy('password', [
            'username' => $request->input('username'),
            'password' => $request->input('password'),
        ]);
    }
}
