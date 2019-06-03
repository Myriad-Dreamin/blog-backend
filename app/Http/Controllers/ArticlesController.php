<?php

namespace App\Http\Controllers;

use Illuminate\Http\Request;
use \App\Article as Article;


class ArticlesController extends Controller
{
    //

    public function index()
    {
        $articles = Article::all();
        return view("articles.index", compact("articles"));
    }
}
