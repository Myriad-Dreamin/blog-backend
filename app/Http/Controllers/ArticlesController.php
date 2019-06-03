<?php

namespace App\Http\Controllers;

use Illuminate\Http\Request;
use \App\Article as Article;
// use \App\Libs\MarkDownX\MarkDownX as MarkDown;

class ArticlesController extends Controller
{
    //

    public function index()
    {
        $articles = Article::all();
        return view("articles.index", compact("articles"));
    }
    
    public function show($id)
    {
        $article = Article::findOrFail($id);
        $local_path = $article->category. '/' . $article->filepath . '.md';
        if(\Storage::disk('articles')->exists($local_path) == FALSE) {
            dd('md file '. $local_path .' not found.');
            // abort(404, 'md file not found.');
        }
        $parser = new \MarkJaxParser;
        $content = $parser->makeHtml(\Storage::disk('articles')->get($local_path));
        return view("articles.article", compact("article", "content"));
    }
}
