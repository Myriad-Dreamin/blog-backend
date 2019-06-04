<?php

namespace App;

use Illuminate\Database\Eloquent\Model;

class Article extends Model
{
    protected $fillable = ['title', 'intro', 'published_at', 'category', 'filepath'];

    protected $dates = ['published_at'];

    // public function setPublishedAtAttribute($date)

    public function scopePublished($query)
    {
        $query->where('published_at', '<=', \Carbon\Carbon::now());
    }
}
