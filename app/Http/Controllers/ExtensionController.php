<?php

namespace App\Http\Controllers;

use Illuminate\Http\Request;

class ExtensionController extends Controller
{
    public function secretToMe()
    {
        return view("extensions.secretlove");
    }

    public function chocolate()
    {
        $chocos = [
            (object) [
                'title' => 'Choco Giving',
                'category' => 'Any',
                'intro' => '要不要也给我推荐一点强大的工具QwQ',
                'ref' => 'http://myriaddreamin.com/secretlove'
            ],
            (object) [
                'title' => 'Git',
                'category' => 'Code',
                'intro' => 'Git包管理工具真的很好用呀QwQ',
                'ref' => 'https://git-scm.com/'
            ],
            (object) [
                'title' => 'Git Desktop',
                'category' => 'Code',
                'intro' => '用过Windows10版本的Git GUI吗',
                'ref' => 'https://desktop.github.com/'
            ],
            (object) [
                'title' => 'Visual Studio Code',
                'category' => 'Code',
                'intro' => '我身边的同学们都在用',
                'ref' => 'https://code.visualstudio.com/'
            ],
            (object) [
                'title' => 'php7',
                'category' => 'Language',
                'intro' => '世界上最好的语言',
                'ref' => 'https://www.php.net/'
            ],
            (object) [
                'title' => 'Python3',
                'category' => 'Language',
                'intro' => '世界上第二好的语言',
                'ref' => 'https://www.python.org/'
            ],
            (object) [
                'title' => 'Matplotlib',
                'category' => 'Pacakge',
                'intro' => '图形绘画工具',
                'ref' => 'https://matplotlib.org/'
            ]
        ];
        return view("extensions.chocolate", compact('chocos'));
    }
}
