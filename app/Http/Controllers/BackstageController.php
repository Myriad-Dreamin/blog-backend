<?php

namespace App\Http\Controllers;

use Illuminate\Http\Request;

class BackstageController extends Controller
{
    //
    public function index()
    {
        return view("backstage.index");
    }
}
