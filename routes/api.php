<?php

use Illuminate\Http\Request;
use \App\Article as Article;

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

Route::middleware('api')->get('/articles', function () {
    $articles = Article::latest()->published()->get();
    return response()->json($articles);
});

Route::middleware('api')->get('/article/{id}', function ($id) {
    $article = Article::findOrFail($id);
    $local_path = $article->category. '/' . $article->filepath . '.md';
    if(\Storage::disk('articles')->exists($local_path) == FALSE) {
        dd('md file '. $local_path .' not found.');
        // abort(404, 'md file not found.');
    }
    $parser = new \MarkJaxParser;
    $parser->enableMathJax(true);
    $article->content = $parser->makeHtml(\Storage::disk('articles')->get($local_path));
    return response()->json($article);
});
